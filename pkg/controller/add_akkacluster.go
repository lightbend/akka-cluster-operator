package controller

import (
	"github.com/lightbend/akka-cluster-operator/pkg/controller/akkacluster"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, akkacluster.Add)
}
