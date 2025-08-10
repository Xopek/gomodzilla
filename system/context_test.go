package system_test

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/Xopek/gomodzilla/system"
)

func TestSystemContextIsCanceledOnSIGTERM(t *testing.T) {
	// Arrange
	testTimeout := time.Second * 3
	isContextCancelled := false

	// Act
	ctx := system.NewSystemContext(context.Background(), 0)
	go func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	// Assert
	select {
	case <-ctx.Done():
		isContextCancelled = true
	case <-time.After(testTimeout):
		break
	}

	if !isContextCancelled {
		t.Error("system context expected to be cancelled on SIGTERM, but it was not")
	}
}

func TestSystemContextIsCanceledOnSIGINT(t *testing.T) {
	// Arrange
	testTimeout := time.Second
	isContextCancelled := false

	// Act
	ctx := system.NewSystemContext(context.Background(), 0)
	go func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// Assert
	select {
	case <-ctx.Done():
		isContextCancelled = true
	case <-time.After(testTimeout):
		break
	}

	if !isContextCancelled {
		t.Error("system context expected to be cancelled on SIGINT, but it was not")
	}
}

func TestSystemContextCallbackIsCalled(t *testing.T) {
	// Arrange
	testTimeout := time.Second * 3
	callbackSignalCh := make(chan string, 1)
	expectedSignal := syscall.SIGTERM

	// Act
	system.NewSystemContext(context.Background(), 0, func(_ context.Context, signal os.Signal) {
		callbackSignalCh <- signal.String()
	})
	go func() {
		syscall.Kill(syscall.Getpid(), expectedSignal)
	}()

	// Assert
	select {
	case signalReceived := <-callbackSignalCh:
		if signalReceived != expectedSignal.String() {
			t.Errorf(
				"system context callback expected to receive signal %q, but got %q",
				expectedSignal.String(),
				signalReceived,
			)
		}
	case <-time.After(testTimeout):
		t.Error("system context callback expected to be called on any received signal, but it was not")
	}
}
