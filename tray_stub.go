//go:build !linux

package main

// startTray — заглушка для платформ без поддержки трея. Приложение работает
// без индикатора.
func startTray(cfg Config, loc Locale) *TrayManager {
	return nil
}