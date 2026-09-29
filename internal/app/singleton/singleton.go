package singleton

import "sync"

const currentUserID = 1

var (
	once        sync.Once
	currentUser *CurrentUser
)

type CurrentUser struct {
	ID uint
}

func GetCurrentUser() *CurrentUser {
	once.Do(func() {
		currentUser = &CurrentUser{ID: currentUserID}
	})

	return currentUser
}
