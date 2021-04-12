package errors

type ErrorAlreadyRegistered struct {}

func (e ErrorAlreadyRegistered) Error() string {
	return "user with such identity already exists"
}

type ErrorNoStage struct {}

func (e ErrorNoStage) Error() string {
	return "NoStageForIdentity"
}
