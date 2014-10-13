/*
 * Spreed WebRTC.
 * Copyright (C) 2013-2014 struktur AG
 *
 * This file is part of Spreed WebRTC.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"testing"
)

const (
	testAppVersion string = "0.0.0+unittests"
)

type fakeClient struct {
}

func (fake *fakeClient) Send(_ Buffer) {
}

func (fake *fakeClient) Reply(_ string, _ interface{}) {
}

type fakeRoomManager struct {
	disallowJoin bool
	joinedRoomID string
	leftRoomID   string
	roomUsers    []*DataSession
	joinedID     string
	leftID       string
	broadcasts   []interface{}
}

func (fake *fakeRoomManager) CanJoinRoom(roomID string) bool {
	return !fake.disallowJoin
}

func (fake *fakeRoomManager) RoomUsers(session *Session) []*DataSession {
	return fake.roomUsers
}

func (fake *fakeRoomManager) JoinRoom(session *Session, _ Sender) {
	fake.joinedID = session.Roomid
}

func (fake *fakeRoomManager) LeaveRoom(session *Session) {
	fake.leftID = session.Roomid
}

func (fake *fakeRoomManager) Broadcast(_ *Session, msg interface{}) {
	fake.broadcasts = append(fake.broadcasts, msg)
}

func NewTestChannellingAPI() (ChannellingAPI, *fakeClient, *Session, *fakeRoomManager) {
	client, roomManager, session := &fakeClient{}, &fakeRoomManager{}, &Session{}
	return NewChannellingAPI(testAppVersion, nil, roomManager, nil, nil, nil, nil, nil, nil, roomManager, nil), client, session, roomManager
}

func Test_ChannellingAPI_OnIncoming_HelloMessage_JoinsTheSelectedRoom(t *testing.T) {
	roomID, ua := "foobar", "unit tests"
	api, client, session, roomManager := NewTestChannellingAPI()

	api.OnIncoming(client, session, &DataIncoming{Type: "Hello", Hello: &DataHello{Id: roomID, Ua: ua}})

	if roomManager.joinedID != roomID {
		t.Errorf("Expected to have joined room %v, but got %v", roomID, roomManager.joinedID)
	}

	if broadcastCount := len(roomManager.broadcasts); broadcastCount != 1 {
		t.Fatalf("Expected 1 broadcast, but got %d", broadcastCount)
	}

	dataSession, ok := roomManager.broadcasts[0].(*DataSession)
	if !ok {
		t.Fatal("Expected a session data broadcast")
	}

	if dataSession.Ua != ua {
		t.Errorf("Expected to have broadcasted a user agent of %v, but was %v", ua, dataSession.Ua)
	}
}

func Test_ChannellingAPI_OnIncoming_HelloMessage_LeavesAnyPreviouslyJoinedRooms(t *testing.T) {
	roomID := "foobar"
	api, client, session, roomManager := NewTestChannellingAPI()

	api.OnIncoming(client, session, &DataIncoming{Type: "Hello", Hello: &DataHello{Id: roomID}})
	api.OnIncoming(client, session, &DataIncoming{Type: "Hello", Hello: &DataHello{Id: "baz"}})

	if roomManager.leftID != roomID {
		t.Errorf("Expected to have left room %v, but got %v", roomID, roomManager.leftID)
	}

	if broadcastCount := len(roomManager.broadcasts); broadcastCount != 3 {
		t.Fatalf("Expected 3 broadcasts, but got %d", broadcastCount)
	}

	dataSession, ok := roomManager.broadcasts[1].(*DataSession)
	if !ok {
		t.Fatal("Expected a session data broadcast")
	}

	if status := "soft"; dataSession.Status != status {
		t.Errorf("Expected to have broadcast a leave status of of %v, but was %v", status, dataSession.Status)
	}
}

func Test_ChannellingAPI_OnIncoming_HelloMessage_DoesNotJoinIfNotPermitted(t *testing.T) {
	api, client, session, roomManager := NewTestChannellingAPI()
	roomManager.disallowJoin = true

	api.OnIncoming(client, session, &DataIncoming{Type: "Hello", Hello: &DataHello{}})

	if broadcastCount := len(roomManager.broadcasts); broadcastCount != 0 {
		t.Fatalf("Expected no broadcasts, but got %d", broadcastCount)
	}
}
