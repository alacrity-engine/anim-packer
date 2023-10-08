package main

import "io/fs"

type FileTracker struct {
	EntryPath string
	Entry     fs.DirEntry
}
