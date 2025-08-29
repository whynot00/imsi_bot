package reader

import (
	"encoding/csv"
	"io"
)

func Load(file io.Reader, lineCh chan<- []string) error {

	r := csv.NewReader(file)

	r.Comma = ';'
	r.LazyQuotes = true

	for {

		line, err := r.Read()
		if err == io.EOF {
			break
		}

		// if line[models.IMSI] == "250203909046342" {
		// 	fmt.Println(line)
		// }

		lineCh <- line

	}

	return nil
}
