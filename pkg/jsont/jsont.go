package jsont

type User struct {
	ID    int64  `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Age   uint8  `json:"age,omitempty"`
}
