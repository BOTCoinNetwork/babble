package mobile

/*
These types are exported and need to be implemented and used by the mobile
application.
*/

//------------------------------------------------------------------------------

// CommitHandler ...
type CommitHandler interface {
	OnCommit(block []byte) (processedBlock []byte)
}

// ExceptionHandler ...
type ExceptionHandler interface {
	OnException(string)
}
