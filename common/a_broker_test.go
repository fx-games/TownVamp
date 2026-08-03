package common

import "testing"

type fakeAction struct {
	BaseAction
	drop bool
}

func (f *fakeAction) Execute(ActionContext) []Action { return nil }
func (f *fakeAction) GetAnimations() []Animation     { return nil }

// More queued batches than actions in any one batch is the shape that used to
// panic: the filter indexed a batch's actions with the batch counter.
func TestClearQueuedActionsWithManyBatches(t *testing.T) {
	broker := &Broker{}
	for i := 0; i < 30; i++ {
		broker.queuedActions = append(broker.queuedActions, []Action{&fakeAction{drop: i%2 == 0}})
	}

	broker.ClearQueuedActions(func(action Action) bool {
		return action.(*fakeAction).drop
	})

	if len(broker.queuedActions) != 15 {
		t.Errorf("expected 15 batches to survive, got %d", len(broker.queuedActions))
	}
	for _, batch := range broker.queuedActions {
		for _, action := range batch {
			if action.(*fakeAction).drop {
				t.Error("an action that should have been dropped survived")
			}
		}
	}
}

func TestClearQueuedActionsCanEmptyEverything(t *testing.T) {
	broker := &Broker{}
	for i := 0; i < 5; i++ {
		broker.queuedActions = append(broker.queuedActions, []Action{&fakeAction{drop: true}})
	}
	broker.ClearQueuedActions(func(Action) bool { return true })
	if len(broker.queuedActions) != 0 {
		t.Errorf("expected everything cleared, %d batches remain", len(broker.queuedActions))
	}
}
