package config

import (
	"firestore-utils/internal/controller"
	"firestore-utils/internal/repository"
)

func GetControllersInstance() GinController {

	return GinController{Controllers: []ControllerRunnable{
		controller.NewController(getGenericRepository()),
	}}
}

func getGenericRepository() *repository.GenericRepository {
	return repository.NewRepository(GetFirestoreClient())
}
