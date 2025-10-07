package store

type Vector struct {
	ID 				string 						`json:"id"`
	Values 		[]float64 				`json:"vector"`
	Metadata 	map[string]string	`json:"metadata"`
}
