package sshtunnel

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Repository — конкретная реализация репозитория
type Repository struct {
	localUsername string
	localHost     string
	tunnelAddress string
}

func NewRepository(localUsername, localHost, tunnelAddress string) *Repository {
	// Конкретные параметры команды:
	//  ssh -D 192.168.64.8:8080 -N nikolai@127.0.0.1
	return &Repository{
		localUsername: localUsername,
		localHost:     localHost,
		tunnelAddress: tunnelAddress,
	}
}

// StartTunnel запускает ssh-туннель в фоне и сохраняет PID.
// Возвращает этот PID, либо ошибку.
func (r *Repository) StartTunnel(privateKeyPath string) (int, error) {

	localhost := fmt.Sprintf("%s@%s", r.localUsername, r.localHost)
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", privateKeyPath, "-D", r.tunnelAddress, "-N", localhost)

	// Запускаем в фоне
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("ошибка запуска ssh: %w", err)
	}

	// Сохраняем PID, если понадобятся последующие операции
	pid := cmd.Process.Pid

	// cmd.Wait() в горутине, чтобы процесс не остался «зомби», когда завершится
	go func() {
		_ = cmd.Wait()
	}()

	return pid, nil
}

// IsRunning проверяет, жив ли процесс с указанным PID.
func (r *Repository) IsRunning(pid int) (bool, error) {
	if pid <= 0 {
		return false, errors.New("невалидный PID")
	}
	err := syscall.Kill(pid, syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, syscall.EPERM) {
		return true, nil
	}
	return false, nil
}

// GetPID ищет процесс ssh -D 192.168.64.8:8080 -N nikolai@127.0.0.1 через ps,
// возвращает PID, если найден. Если не найден, вернёт ошибку.
func (r *Repository) GetPID() (int, bool, error) {
	// Выполним ps aux и поищем строку "ssh -D 192.168.64.8:8080 -N nikolai@127.0.0.1"
	// Можно использовать pgrep, но для демонстрации — разбор ps.
	cmd := exec.Command("ps", "aux")
	out, err := cmd.Output()
	if err != nil {
		return 0, false, fmt.Errorf("не удалось выполнить ps: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	localhost := fmt.Sprintf("%s@%s", r.localUsername, r.localHost)
	args := []string{"-D", r.tunnelAddress, "-N", localhost}
	searchLine := strings.Join(args, " ") // "-D 192.168.64.8:8080 -N nikolai@127.0.0.1"
	// Важно: первая часть "ssh" в ps может отображаться иначе (полный путь /usr/bin/ssh).
	// Можно проверять подстроку без "ssh", а только "-D 192.168.64.8:8080 -N ..."
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, searchLine) {
			// парсим PID (оно второе поле в ps aux? завист от ОС)
			fields := strings.Fields(line)
			//  USER       PID  %CPU ...
			// fields[1] -> PID
			pidStr := fields[1]
			pid, _ := strconv.Atoi(pidStr)
			if pid > 0 {
				return pid, true, nil
			}
		}
	}

	return 0, false, nil
}

// StopTunnel посылает сигнал SIGTERM (или SIGKILL) процессу по PID
func (r *Repository) StopTunnel(pid int) error {
	if pid <= 0 {
		return errors.New("невалидный PID для остановки")
	}
	err := syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("не удалось остановить процесс %d: %w", pid, err)
	}
	return nil
}

func (r *Repository) CheckHealth() (bool, error) {
	proxyURL, err := url.Parse(fmt.Sprintf("socks5://%s", r.tunnelAddress))
	if err != nil {
		return false, err
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get("https://ya.ru")
	if err != nil {
		return false, fmt.Errorf("запрос через туннель не удался: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

func (r *Repository) CheckHealthTCP() (bool, error) {
	// Создаём SOCKS5-прокси-дилер
	dialer, err := proxy.SOCKS5("tcp", r.tunnelAddress, nil, proxy.Direct)
	if err != nil {
		return false, fmt.Errorf("создание SOCKS5 dialer: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := dialer.(proxy.ContextDialer).DialContext(ctx, "tcp", "8.8.8.8:53")
	if err != nil {
		return false, nil
	}
	conn.Close()
	return true, nil
}
