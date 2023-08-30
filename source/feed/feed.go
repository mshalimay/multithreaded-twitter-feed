// A thread-safe feed implemented as a linked list with a coarse-grained strategy 
// using a read-write lock.

package feed

import (
	"proj2/lock"
)

//Feed represents a user's twitter feed
// @Add: inserts a new post to the feed
// @Remove: deletes the post with the given timestamp
// @Contains: determines whether a post with the given timestamp is inside a feed
// @ReturnFeed: returns the whole feed as a slice of Post structs
type Feed interface {
	Add(body string, timestamp float64)
	Remove(timestamp float64) bool
	Contains(timestamp float64) bool
	ReturnFeed() []Post
}

//feed is the internal representation of a user's twitter feed (hidden from outside packages)
type feed struct {
	start 		*post 			// a pointer to the beginning post
	rwLock 		lock.RWLock		// a read-write lock
}

//post is the internal representation of a post on a user's twitter feed (hidden from outside packages)
type post struct {
	body      string 		// the text of the post
	timestamp float64  		// Unix timestamp of the post
	next      *post  		// the next post in the feed
	content   *Post			// helper struct for returning the feed
	removed   bool			// flag to indicate if post was deleted (used in `feed2.go` for optimistic locking)
}

// Post is a helper struct for returning the feed, containing only the body and timestamp of a feed post
// i.e., does not contain the elements from `post` that should not be exposed (eg: pointers to implementation of linked list)
// Obs: having this struct minimizes the work when returning the whole feed
type Post struct {
	Body      *string 		`json:"body"`	    // the text of the post
	Timestamp *float64  	`json:"timestamp"`	// Unix timestamp of the post
}

//NewPost creates and returns a new post value given its body and timestamp
func newPost(body string, timestamp float64, next *post) *post {
	p := &post{body, timestamp, next, nil, false}
	p.content = &Post{Body: &p.body, Timestamp: &p.timestamp}
	return p
}

//NewFeed creates a empty user feed and returns a pointer to it
func NewFeed() Feed {
	rwLock := lock.NewRWLock()
	// rwLock := lock.NewRWLockFaster()
	return &feed{start: nil, rwLock: rwLock}
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc. You may need to insert a new post somewhere in the feed because
// the given timestamp may not be the most recent.
func (f *feed) Add(body string, timestamp float64) {
	// creates a new post/node with the given body and timestamp
	newPost := newPost(body, timestamp, nil)

	// get a writer lock to update the feed
	// Obs1: taking a writer lock here avoid other threads to read/update the feed 
	// 		 while another thread is updating it.
	// Obs2: this is not the most performant implementation; see `feed2.go` for
	// 		 a better one in this regard.

	f.rwLock.Lock()
	defer f.rwLock.Unlock()

	// iterates over all feed; if timestamp in the middle add post to the middle of the feed
	curPost := f.start
	// if feed is empty or post is the most recent, add post to the beginning of the feed
	if curPost == nil || timestamp > curPost.timestamp {
		newPost.next = curPost
		f.start = newPost	
		return
	// else, traverse the feed until find the correct position to insert the post	
	} else {
		for curPost.next != nil && timestamp < curPost.next.timestamp{
			curPost = curPost.next
		}
		newPost.next = curPost.next
		curPost.next = newPost
		return
	}
}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *feed) Remove(timestamp float64) bool {

	// get a writer lock to update the feed
	// see obs in Add() for more details
	f.rwLock.Lock()
	defer f.rwLock.Unlock()

	
	// iterate over all feed; if timestamp in the middle remove post and update pointers
	// such that: old feed: a -> b -> c ===> new feed: a -> c	
	curPost := f.start
	
	// if feed is empty, return false
	if curPost == nil{
		return false
	// if post to be removed is the most recent, remove it and update feed
	} else if timestamp == curPost.timestamp{
		f.start = curPost.next
		return true
	// else, traverse the feed until find the correct position to remove the post
	} else {

		for curPost.next != nil && timestamp != curPost.next.timestamp{
			curPost = curPost.next
		}
		// if next post is nil, end of feed was reached and post was not found
		if curPost.next == nil{
			return false
		// else, remove post and update pointers (e.g. old feed: a -> b -> c ===> new feed: a -> c)
		} else {
			curPost.next = curPost.next.next
			return true
		}
	}
}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Contains(timestamp float64) bool {
	
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
func (f *feed) ReturnFeed() []Post {
	var feed []Post

	f.rwLock.RLock()
	defer f.rwLock.RUnlock()
	curPost := f.start
	for curPost != nil {
		feed = append(feed, *curPost.content)
		curPost = curPost.next
	}
	return feed
}