package wsmodel

type Reaction struct {
	MessageType string
	MessageID   int
	Reaction    bool
}

func (r *Reaction) Validate() string {
	if isEmpty(r.MessageType) {
		return "type of message missing"
	}
	if r.MessageID<=0 {
		return "wrong message id missing"
	}
	return ""
}