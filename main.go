package main

func main() {
	if err := run(); err != nil {
		LogError("Failed to run", "実行に失敗しました", "error", err)
	}
}

func run() error {
	return nil
}
