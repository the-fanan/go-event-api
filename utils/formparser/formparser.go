package formparser

import (
	"strings"
	"mime/multipart"
)

type NestedFormData struct {
	Value *ValueNode
	File *FileNode
}

type ValueNode struct {
	Value []string
	Children map[string]*ValueNode
}

type FileNode struct {
	Value []*multipart.FileHeader
	Children map[string]*FileNode
}

func (fd *NestedFormData) ParseValues(m map[string][]string){
	n := &ValueNode{
		Children: make(map[string]*ValueNode),
	}
	for key, val := range m {
		keys := strings.Split(key,".")
		fd.nestValues(n, &keys, val)
	}
	fd.Value = n
}

func (fd *NestedFormData) ParseFiles(m map[string][]*multipart.FileHeader){
	n := &FileNode{
		Children: make(map[string]*FileNode),
	}
	for key, val := range m {
		keys := strings.Split(key,".")
		fd.nestFiles(n, &keys, val)
	}
	fd.File = n
}

func (fd *NestedFormData) nestValues(n *ValueNode, k *[]string, v []string) {
	var key string
	key, *k = (*k)[0], (*k)[1:]
	if len(*k) == 0 {
			if _, ok := n.Children[key]; ok {
					n.Children[key].Value = append(n.Children[key].Value, v...)
			} else {
					cn := &ValueNode{
							Value: v,
							Children: make(map[string]*ValueNode),
					}
					n.Children[key] = cn
			}
	} else {
		if _, ok := n.Children[key]; ok {
			fd.nestValues(n.Children[key], k,v)
		} else {
			cn := &ValueNode{
				Children: make(map[string]*ValueNode),
			}
			n.Children[key] = cn
			fd.nestValues(cn, k,v)
		}
	}
}

func (fd *NestedFormData) nestFiles(n *FileNode, k *[]string, v []*multipart.FileHeader){
	var key string
	key, *k = (*k)[0], (*k)[1:]
	if len(*k) == 0 {
		if _, ok := n.Children[key]; ok {
			n.Children[key].Value = append(n.Children[key].Value, v...)
		} else {
			cn := &FileNode{
				Value: v,
				Children: make(map[string]*FileNode),
			}
			n.Children[key] = cn
		}
	} else {
		if _, ok := n.Children[key]; ok {
			fd.nestFiles(n.Children[key], k,v)
		} else {
			cn := &FileNode{
				Children: make(map[string]*FileNode),
			}
			n.Children[key] = cn
			fd.nestFiles(cn, k,v)
		}
	}
}