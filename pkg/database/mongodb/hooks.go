package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// CreatingHook call before saving new document into database
type CreatingHook interface {
	Creating() error
}

// CreatedHook call after document has been created
type CreatedHook interface {
	Created() error
}

// UpdatingHook call when before updating document
type UpdatingHook interface {
	Updating() error
}

// UpdatedHook call after document updated
type UpdatedHook interface {
	Updated(result *mongo.UpdateResult) error
}

// SavingHook call before save document(new or existed
// document) into database.
type SavingHook interface {
	Saving() error
}

// SavedHook call after document has been saved in database.
type SavedHook interface {
	Saved() error
}

// DeletingHook call before deleting document
type DeletingHook interface {
	Deleting() error
}

// DeletedHook call after document has been deleted)
type DeletedHook interface {
	Deleted(result *mongo.DeleteResult) error
}

func callToBeforeCreateHooks(document Document) error {
	if hook, ok := document.(CreatingHook); ok {
		if err := hook.Creating(); err != nil {
			return err
		}
	}

	if hook, ok := document.(SavingHook); ok {
		if err := hook.Saving(); err != nil {
			return err
		}
	}

	return nil
}

func callToBeforeUpdateHooks(document Document) error {
	if hook, ok := document.(UpdatingHook); ok {
		if err := hook.Updating(); err != nil {
			return err
		}
	}

	if hook, ok := document.(SavingHook); ok {
		if err := hook.Saving(); err != nil {
			return err
		}
	}

	return nil
}

func callToAfterCreateHooks(document Document) error {
	if hook, ok := document.(CreatedHook); ok {
		if err := hook.Created(); err != nil {
			return err
		}
	}

	if hook, ok := document.(SavedHook); ok {
		if err := hook.Saved(); err != nil {
			return err
		}
	}

	return nil
}

func callToAfterUpdateHooks(updateResult *mongo.UpdateResult, document Document) error {
	if hook, ok := document.(UpdatedHook); ok {
		if err := hook.Updated(updateResult); err != nil {
			return err
		}
	}

	if hook, ok := document.(SavedHook); ok {
		if err := hook.Saved(); err != nil {
			return err
		}
	}

	return nil
}

func callToBeforeDeleteHooks(document Document) error {
	if hook, ok := document.(DeletingHook); ok {
		if err := hook.Deleting(); err != nil {
			return err
		}
	}

	return nil
}

func callToAfterDeleteHooks(deleteResult *mongo.DeleteResult, document Document) error {
	if hook, ok := document.(DeletedHook); ok {
		if err := hook.Deleted(deleteResult); err != nil {
			return err
		}
	}

	return nil
}