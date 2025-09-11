package handlers

import (
	"encoding/json"
	"os"
)

func saveToFile(form *ContactForm) error {
	// Превращаем структуру в красивый JSON
	data, err := json.MarshalIndent(form, "", "  ")
	if err != nil {
		return err
	}

	// Открываем файл на запись. Если файла нет - он создается.
	// Флаг os.O_APPEND означает, что мы дописываем в конец файла.
	// Флаг os.O_CREATE означает создать файл, если его нет.
	// Флаг os.O_WRONLY означает только запись.
	// Права 0644 - стандартные права для файла.
	file, err := os.OpenFile("/Users/dimsanyc/Documents/goPr/http/bd/jsonsWithData.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close() // Гарантированно закрываем файл при выходе из функции

	// Добавляем запись в файл с таймстампом и запятой для корректного JSON массива
	_, err = file.WriteString(string(data) + ",\n")
	if err != nil {
		return err
	}

	return nil
}
