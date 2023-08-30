// A thread-safe feed implemented as a linked list with an optimistic strategy using a read-write lock.
// Differences to the implementation in `feed`: threads acquire READ locks when traversing the feed
// and only acquire a WRITE lock when updating the feed. If the feed changes while changing locks, 
// the thread retries the operation.

package feed

import (
	"proj2/lock"
)


//feed is the internal representation of a user's twitter feed (hidden from outside packages)
type optFeed struct {
	start 		*post 			// a pointer to the beginning post
	rwLock 		lock.RWLock		// a read-write lock
}

//NewFeed creates a empty user feed and returns a pointer to it
func NewOptFeed() Feed {
	rwLock := lock.NewRWLock()
	// rwLock := lock.NewRWLockFaster()
	sentinelPost := newPost("", -1, nil)
	return &optFeed{start: sentinelPost, rwLock: rwLock}
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc. You may need to insert a new post somewhere in the feed because
// the given timestamp may not be the most recent.
func (f *optFeed) Add(body string, timestamp float64) {
	// creates a new post/node with the given body and timestamp
	newPost := newPost(body, timestamp, nil)

	for {
		// Acquire a read lock to traverse feed
		f.rwLock.RLock()
		// if feed is empty or post is the most recent, add post to the beginning of the feed (*)
		if f.start.next == nil || timestamp > f.start.next.timestamp {
			f.rwLock.RUnlock()
			// acquire a write lock to try updating feed
			f.rwLock.Lock()
			// check if condition (*) still holds; if so, update feed; else, release lock and retry
			if f.start.next == nil || timestamp > f.start.next.timestamp {
				newPost.next = f.start.next
				f.start.next = newPost
				f.rwLock.Unlock()
				return
			} else {
				f.rwLock.Unlock()
				continue
			}
		// else, traverse the feed until find the correct position to insert the post	
		} else {

			curPost := f.start.next
			for curPost.next != nil && timestamp < curPost.next.timestamp{
				curPost = curPost.next
			}
			// insertion point found => release read lock and acquire write lock to update feed
			f.rwLock.RUnlock()
			f.rwLock.Lock()

			// check if in the change of locks, insertion point is still valid; 
			// if so, update feed; else, release lock and retry
			// Explanation of the conditions. E.g.: feed = 12 -> 10 -> 3 ; new post: 5
			// - condition 1 checks if 10 not deleted; in this case we need to change the pointer of 12 not 10, so retry.
			// - condition 2 checks if the insertion point is still correct. This might not be the case if another 
			//   thread inserts a 6 resulting in feed = 12 -> 10 -> 6 -> 3; continuing with the operation would result
			//   in feed = 12 -> 10 -> 5 -> 6 -> 3 so retry. (i.e., curPost is lagging, must update to 6)
			if !curPost.removed && (curPost.next == nil || timestamp > curPost.next.timestamp) {
				newPost.next = curPost.next
				curPost.next = newPost
				f.rwLock.Unlock()
				return
			} else {
				f.rwLock.Unlock()
				continue
			}
		}
	}
}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *optFeed) Remove(timestamp float64) bool {

	// iterate over all feed; if timestamp in the middle remove post and update pointers
	// such that: old feed: a -> b -> c ===> new feed: a -> c	

	for{
		// Acquire a read lock to traverse feed
		f.rwLock.RLock()

		// if feed is empty, return false
		if f.start.next == nil {
			f.rwLock.RUnlock()
			return false
		
		// if post to be removed is the most recent, remove it and update feed
		} else if timestamp == f.start.next.timestamp {
			// release read lock and acquire write lock to update feed
			f.rwLock.RUnlock()
			f.rwLock.Lock()
			// check if condition still holds; if so, update feed; else, release lock and retry
			if  timestamp == f.start.next.timestamp {
				f.start.next.removed = true
				f.start.next = f.start.next.next
				f.rwLock.Unlock()
				return true
			} else {
				f.rwLock.Unlock()
				continue
			}
			
		// else, traverse the feed until find the correct position to remove the post
		} else {
			curPost := f.start.next
			for curPost.next != nil && timestamp != curPost.next.timestamp{
				curPost = curPost.next
			}
			// if post to be removed is not in the feed, return false
			f.rwLock.RUnlock()
			f.rwLock.Lock()
			// check if removing point still valid; if so, update feed; else, release lock and retry
			if !curPost.removed && (curPost.next == nil || timestamp == curPost.next.timestamp) {
				// if next post is still nil, end of feed was reached and post was not found
				if curPost.next == nil{
					f.rwLock.Unlock()
					return false
				// else, remove post, update pointers and annotate post as removed
				// (e.g. old feed: curPost -> next -> next.next => curPost -> next.next)
				} else {
					// annotate post as removed. This is needed because the node will be dangling and
					// other threads would still be able to use it to find next nodes and think
					// they are doing a valid operation.
					curPost.next.removed = true
					curPost.next = curPost.next.next
					f.rwLock.Unlock()
					return true
				}
			} else {
				f.rwLock.Unlock()
				continue
			}
		}
	}
}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *optFeed) Contains(timestamp float64) bool {
	// get a reader lock to read the feed
	// Obs: this assumes we take a 'snapshot' of the feed when contains is called and return true/false based on it. 
	// E.g.: if at time `t` `contains` is called and `A` is not in the feed, it will return false even if 
	// a concurrent thread is trying to add `A` to the feed. 
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()

	// iterate over all feed; if found timestamp, return true
	curPost := f.start	
	for curPost != nil {
		if curPost.timestamp == timestamp{
			return true
		}
		curPost = curPost.next
	}
	return false
}

// ReturnFeed returns the whole feed as a slice of Post structs
func (f *optFeed) ReturnFeed() []Post {
	// get a reader lock to read the feed
	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	var feed []Post
	curPost := f.start.next
	for curPost != nil {
		feed = append(feed, *curPost.content)
		curPost = curPost.next
	}
	return feed
} 