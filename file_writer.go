package tracker

import (
	"fmt"
	"os"
	"path"
)

type FileWriter struct {
	messages     chan Message
	torrent      *TorrentInfo
	torrent_file string
	files        []filePath
	root         string
}

type filePath struct {
	dir string
	f   *os.File
}

func NewFileWriter(t *TorrentInfo, tf string) *FileWriter {
	fw := new(FileWriter)
	fw.messages = make(chan Message)
	fw.torrent = t
	fw.torrent_file = tf
	return fw
}

func (fw *FileWriter) Run() {
	files := make([]filePath, 0)

	root := path.Dir(fw.torrent_file)

	if !path.IsAbs(root) {
		cwd, _ := os.Getwd()
		root = path.Join(cwd, root)
	}
	err := os.Chdir(root)
	if err != nil {
		fmt.Println(err)
		// handle this somehow
	}
	if fw.torrent.numfiles > 1 {
		root = path.Join(root, fw.torrent.name)
		err = os.MkdirAll(root, os.ModeDir)
		if err != nil {
			fmt.Println(err)
			// handle this too
		}

	} else {
		root = path.Join(root, path.Dir(fw.torrent.name))
		err = os.MkdirAll(root, os.ModeDir)
		if err != nil {
			fmt.Println(err)
		}
	}
	os.Chdir(root)
	for _, f := range fw.torrent.files {
		dir := path.Dir(f.path)
		err = os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			fmt.Println(err)
			// handle
		}
		os.Chdir(dir)
		fi, err := os.Create(path.Base(f.path))
		if err != nil {
			fmt.Println(err)
		}
		fp := filePath{dir, fi}
		files = append(files, fp)

		b := make([]byte, f.length)
		_, err = fi.WriteAt(b, 0)
		if err != nil {
			fmt.Println(err)
		}
		os.Chdir(root)
	}

	fw.files = files
	fw.root = root

	for m := range fw.messages {
		switch m.kind() {
		case i_write_block:
			msg := m.(InternalWriteBlockMessage)
			fw.write(msg.bytes, msg.index)
		}
	}
}

func (fw *FileWriter) write(b []byte, index int) {
	offset := int64(index) * fw.torrent.pieceLength
	remaining := len(b)
	next_byte := 0
	cur := int64(0)
	prev := int64(0)
	os.Chdir(fw.root)
	for i, f := range fw.torrent.files {
		if remaining <= 0 {
			break
		}
		cur += f.length
		if offset < cur {
			file_offset := offset - prev
			to_write := 0
			if int64(remaining) < f.length-file_offset {
				to_write = next_byte + remaining
			} else {
				to_write = next_byte + int(f.length-file_offset)
			}
			os.Chdir(fw.files[i].dir)
			fw.files[i].f.WriteAt(b[next_byte:to_write], file_offset)
			remaining -= to_write - next_byte
			next_byte += to_write
			offset += int64(to_write)
		}
		prev += f.length
		os.Chdir(fw.root)
	}
}
