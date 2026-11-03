package localsend

import "sort"

func sortFiles(files []FileMeta) {
	sort.Slice(files, func(i, j int) bool { return files[i].FileName < files[j].FileName })
}
