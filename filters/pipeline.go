package filters

type Pipeline struct {
	Filters []*Filter
}

var pipelines Pipeline

// func GetPipeline(name string) Pipeline {
// 	if pipe, ok := pipelines[name]; ok {
// 		return pipe
// 	}
// 	return nil
// }
