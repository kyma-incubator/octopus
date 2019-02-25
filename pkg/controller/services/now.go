package services

import (
	"time"
)

type Now struct {
}

func (n *Now) Now() time.Time {
	return time.Now()

}
