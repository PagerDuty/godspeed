// Copyright 2014 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

// Package godspeed is a statsd client for the Datadog extension of statsd
// (DogStatsD). It is used to emit statsd stats as well as the Datadog-specific
// events. This client also has the ability to tag all outgoing statsd metrics.
// Godspeed is meant for synchronous calls, while AsyncGodspeed is used for
// what it says on the tin.
//
// The name godspeed is a bit of a rhyming slang twist on DogStatsD. It's
// also a poke at the fact that the statsd protocol's transport mechanism is UDP...
//
// DogStatsD is a copyright of Datadog <info@datadoghq.com>
package godspeed

import "net"

const (
	// DEFAULT_HOST is 127.0.0.1 (localhost)
	DEFAULT_HOST = "127.0.0.1"

	// DEFAULT_PORT is 8125
	DEFAULT_PORT uint16 = 8125

	// MAX_BYTES is the largest UDP datagram we will try to send
	// this is 8192 bytes minus the size of a UDP header
	MAX_BYTES = 8192 - 8
)

// Godspeed is an unbuffered Statsd client with compatability geared towards the Datadog statsd format
// It consists of Conn (*net.UDPConn) object for sending metrics over UDP,
// Namespace (string) for namespacing metrics, and Tags ([]string) for tags to send with stats
type Godspeed struct {
	// Conn is the UDP connection used for sending the statsd emissions
	Conn *net.UDPConn

	// Namespace is the namespace all stats emissions are prefixed with:
	// <namespace>.<statname>
	Namespace string

	// Tags is the slice of tags to append to each stat emission
	Tags []string

	// AutoTruncate specifies whether or not we will try to truncate a stat
	// before emitting it or just return an error. This is most helpful when
	// using AsyncGodspeed. However, it can result in invalid stat being emitted
	// due to the body being truncated. Meant for when a single emission would
	// be greater than 8192 bytes.
	AutoTruncate bool
}

// New returns a new instance of a Godspeed statsd client.
// This method takes the host as a string, and port as a uint16.
// There is also the ability for autoTruncate. If your metric is longer than MAX_BYTES
// autoTruncate can be used to truncate the message instead of erroring. This doesn't work
// on events and will always return an error.
func New(host string, port uint16, autoTruncate bool) (g *Godspeed, err error) {
	// build a new UDP dialer
	c, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(host), Port: int(port)})

	// if it failed return a pointer to an empty Godspeed struct, and the error
	if err != nil {
		return nil, err
	}

	// build a new Godspeed struct with the UDPConn
	g = &Godspeed{
		Conn:         c,
		Tags:         make([]string, 0),
		AutoTruncate: autoTruncate,
	}

	return
}

// NewDefault is the same as New() except it uses DEFAULT_HOST and DEFAULT_PORT for the connection.
func NewDefault() (g *Godspeed, err error) {
	g, err = New(DEFAULT_HOST, DEFAULT_PORT, false)
	return
}

// AddTag allows you to add a tag for all future emitted stats.
// It takes the tag as a string, and returns a []string containing all Godspeed tags
func (g *Godspeed) AddTag(tag string) []string {
	// return early if the tag already exists
	for _, v := range g.Tags {
		if tag == v {
			return g.Tags
		}
	}

	// add the tag
	g.Tags = append(g.Tags, tag)

	return g.Tags
}

// AddTags is like AddTag(), except it tages a []string and adds each contained string
// This also returns a []string containing the current tags
func (g *Godspeed) AddTags(tags []string) []string {
	// if we already have tags add each tag one at a time
	// otherwise unique the list and assign it directly
	if len(g.Tags) > 0 {
		for _, tag := range tags {
			g.AddTag(tag)
		}
	} else {
		g.Tags = uniqueTags(tags)
	}

	return g.Tags
}

// SetNamespace allows you to prefix all of your metrics with a certain namespace
func (g *Godspeed) SetNamespace(ns string) {
	g.Namespace = trimReserved(ns)
}
