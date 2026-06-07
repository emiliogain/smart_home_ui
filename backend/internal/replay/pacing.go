package replay

import "time"

// SleepBetweenRows controls wall-clock pacing between replayed rows or batches.
//
// If everySeconds > 0: sleeps exactly that many seconds after each row (fixed UI cadence;
// ignores how far apart timestamps were in the dataset).
//
// Otherwise, if playback > 0 and delta > 0: sleeps delta/playback so playback==1 replays
// in real time relative to dataset timestamps. playback==2 plays 2× faster.
//
// If playback <= 0 and everySeconds <= 0: no sleep.
//
// maxWaitSeconds > 0 caps the sleep from the delta/playback branch only (0 = no cap).
func SleepBetweenRows(delta time.Duration, everySeconds, playback, maxWaitSeconds float64) {
	if everySeconds > 0 {
		time.Sleep(time.Duration(everySeconds * float64(time.Second)))
		return
	}
	if playback <= 0 || delta <= 0 {
		return
	}
	wait := time.Duration(float64(delta) / playback)
	if maxWaitSeconds > 0 {
		cap := time.Duration(maxWaitSeconds * float64(time.Second))
		if wait > cap {
			wait = cap
		}
	}
	time.Sleep(wait)
}
