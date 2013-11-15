package gotp

type Pid struct {
	//the channel over which to receive messages and deliver them to the actor
	recv chan Message
	//the channel to signal that the actor backing this pid should shut down
	stop chan error
	//the channel to signal a watcher that the actor backing this pid errored
	errored chan error
}
