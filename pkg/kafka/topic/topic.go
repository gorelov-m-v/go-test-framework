package topic

type Topic interface {
	Name() string
}

type T[TName TopicName] struct{}

func (t T[TName]) Name() string {
	var name TName
	return string(name)
}

type TopicName interface {
	~string
	TopicName() string
}
