package galax

import "os"

func log(err error) {
	println("[ERROR]", err)

	file := "log.txt"

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err := f.WriteString(err.Error() + "\n"); err != nil {
		panic(err)
	}

	if err := f.Sync(); err != nil {
		panic(err)
	}

	panic(err)
}
