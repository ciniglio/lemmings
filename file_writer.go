package tracker

import (
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
	f   string
}

func NewFileWriter(t *TorrentInfo, tf string) *FileWriter {
	fw := new(FileWriter)
	fw.messages = make(chan Message)
	fw.torrent = t
	fw.torrent_file = tf
	return fw
}

func (fw *FileWriter) Run() {
	var perm os.FileMode = 0731
	var err error
	files := make([]filePath, 0)

	root := path.Dir(fw.torrent_file)

	if !path.IsAbs(root) {
		cwd, err := os.Getwd()
		if err != nil {
			errorl.Println("Error getting wd 0: ", err)
		}
		root = path.Join(cwd, root)
	}
	if fw.torrent.numfiles > 1 {
		root = path.Join(root, fw.torrent.name)
		err = os.MkdirAll(root, os.ModeDir|perm)
		if err != nil {
			errorl.Println("Error Mkdir multifile", err)
			// handle this too
		}

	} else {
		root = path.Join(root, path.Dir(fw.torrent.name))
		err = os.MkdirAll(root, os.ModeDir|perm)
		if err != nil {
			errorl.Println("Error Mkdir 1 file", err)
		}
	}
	if err != nil {
		errorl.Println("Error Chdir root", err)
		// handle this somehow
	}
	for _, f := range fw.torrent.files {
		dir := root
		for i, d := range f.path {
			if i == len(f.path)-1 {
				break
			}
			dir = path.Join(dir, d)
			err = os.MkdirAll(dir, os.ModeDir|perm)
			if err != nil {
				errorl.Println("Error Create Dir", err)
				// handle
			}
		}

		file_name := f.path[len(f.path)-1]
		fi, err := os.Create(path.Join(dir, file_name))
		if err != nil {
			errorl.Println("Error Create file", err, f)
		}
		debugl.Println("DIR:::", dir)
		fp := filePath{dir, file_name}
		files = append(files, fp)

		fi.Close()
		if err != nil {
			errorl.Println("Error Chdir root", err)
		}
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
	var perm os.FileMode = 0666
	offset := int64(index) * fw.torrent.pieceLength
	remaining := len(b)
	next_byte := 0
	cur := int64(0)
	prev := int64(0)
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
			fi, err := os.OpenFile(path.Join(fw.files[i].dir, fw.files[i].f), os.O_RDWR, perm)
			if err != nil {
				errorl.Println("Error opening: ", err, fw.files[i])
			}
			fi.WriteAt(b[next_byte:to_write], file_offset)
			fi.Close()
			remaining -= to_write - next_byte
			next_byte += to_write
			offset += int64(to_write)
		}
		prev += f.length
	}
}
