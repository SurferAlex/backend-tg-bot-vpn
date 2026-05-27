package logtail

import (
	"bufio"
	"io"
	"os"
)

type Tailer struct {
	path     string
	startEOF bool
	file     *os.File
	reader   *bufio.Reader
}

func Open(path string, startEOF bool) (*Tailer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	t := &Tailer{
		path:     path,
		startEOF: startEOF,
		file:     f,
		reader:   bufio.NewReader(f),
	}
	if startEOF {
		if _, err := f.Seek(0, io.SeekEnd); err != nil {
			_ = f.Close()
			return nil, err
		}
		t.reader = bufio.NewReader(f)
	}
	return t, nil
}

func (t *Tailer) Close() error {
	if t.file == nil {
		return nil
	}
	err := t.file.Close()
	t.file = nil
	t.reader = nil
	return err
}

func (t *Tailer) reopen() error {
	if t.file != nil {
		_ = t.file.Close()
	}
	f, err := os.Open(t.path)
	if err != nil {
		return err
	}
	t.file = f
	if t.startEOF {
		if _, err := f.Seek(0, io.SeekEnd); err != nil {
			_ = f.Close()
			t.file = nil
			return err
		}
	}
	t.reader = bufio.NewReader(f)
	return nil
}

func (t *Tailer) ReadLines() ([]string, error) {
	if t.file == nil {
		if err := t.reopen(); err != nil {
			return nil, err
		}
	}

	st, err := t.file.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() == 0 {
		return nil, nil
	}
	cur, err := t.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if st.Size() < cur {
		if err := t.reopen(); err != nil {
			return nil, err
		}
	}

	var lines []string
	for {
		line, err := t.reader.ReadString('\n')
		if len(line) > 0 {
			if line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if line != "" {
				lines = append(lines, line)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return lines, err
		}
	}
	return lines, nil
}
