package sshtunnel

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Repository — конкретная реализация репозитория
type Repository struct {
	localUsername string
}

func NewRepository(localUsername string) *Repository {
	return &Repository{
		localUsername: localUsername,
	}
}

// StartTunnel запускает ssh-туннель в фоне и сохраняет PID.
// Возвращает этот PID, либо ошибку.
func (r *Repository) StartTunnel(privateKeyPath string) (int, error) {

	localhost := fmt.Sprintf("%s@127.0.0.1", r.localUsername)
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-i", privateKeyPath, "-D", "127.0.0.1:8080", "-N", localhost)

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
	localhost := fmt.Sprintf("%s@127.0.01", r.localUsername)
	args := []string{"-D", "127.0.0.1:8080", "-N", localhost}
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
