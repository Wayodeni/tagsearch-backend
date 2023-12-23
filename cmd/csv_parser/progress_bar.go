package main

import (
	"fmt"
	"sync"
	"time"
)

type StateStore struct {
	Processed int
	Left      int
	mu        *sync.RWMutex
}

func NewStateStore(processed, left int) *StateStore {
	return &StateStore{
		Processed: processed,
		Left:      left,
		mu:        &sync.RWMutex{},
	}
}

func (s *StateStore) Set(processed, left int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Processed = processed
	s.Left = left
}

func (s *StateStore) Get() (processed int, left int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Processed, s.Left
}

/*
startTime - time when long running operation was started
currentIndex - which object now is being processed of N (155th out of 1000)
finalIndex - overall quantity of objects (N)
printPeriod - after how much objects info must be printed (print info after every printPeriod objects processed)
*/
type ProgressBar struct {
	startTime   time.Time
	finalIndex  int
	printPeriod int
	stateStore  *StateStore
}

func NewProgressBar(startTime time.Time, finalIndex int, printPeriod int) *ProgressBar {
	pb := &ProgressBar{
		startTime:   startTime,
		finalIndex:  finalIndex,
		printPeriod: printPeriod,
		stateStore:  NewStateStore(0, finalIndex),
	}
	go pb.printProgressBar()
	return pb
}

func (pb *ProgressBar) printProgressBar() {
	for {
		time.Sleep(time.Duration(pb.printPeriod) * time.Millisecond)
		processed, left := pb.stateStore.Get()
		if left == 0 {
			return
		}
		elapsedTime := time.Since(pb.startTime).Seconds()
		additionSpeed := float64(processed) / elapsedTime
		timeLeft := (float64(pb.finalIndex - processed)) / additionSpeed
		secondsLeft := time.Duration(timeLeft) * time.Second
		percentOfObjectsProcessed := (float64(processed) / float64(pb.finalIndex)) * 100
		fmt.Printf("added to db %.0f%% (%d) documents\n", percentOfObjectsProcessed, processed)
		fmt.Printf("time left: '%.0f' hrs or '%.0f' mins or '%.0f' secs\n", secondsLeft.Hours(), secondsLeft.Minutes(), secondsLeft.Seconds())
		fmt.Printf("addition speed is %.f docs/sec. WOW! \n", additionSpeed)
		fmt.Printf("docs left: %d \n", left)
	}
}

func (pb *ProgressBar) Increment() {
	processed, left := pb.stateStore.Get()
	pb.stateStore.Set(processed+1, left-1)
}
