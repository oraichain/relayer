package processor

import (
	"errors"
	"testing"
)

func TestIsUnqueryableIBCEvent(t *testing.T) {
	t.Parallel()
	err := errors.New("no ibc messages found for send_packet query: send_packet.packet_src_channel='channel-216'")
	if !isUnqueryableIBCEvent(err, "send_packet") {
		t.Fatal("expected unqueryable send_packet error")
	}
	if isUnqueryableIBCEvent(err, "write_acknowledgement") {
		t.Fatal("expected different event name to not match")
	}
	if isUnqueryableIBCEvent(nil, "send_packet") {
		t.Fatal("nil error should not match")
	}
}

func TestPrunedPacketBlacklist(t *testing.T) {
	t.Parallel()
	pp := &PathProcessor{prunedPackets: make(map[prunedPacketKey]prunedPacketRecord)}
	key := prunedPacketKey{SourceChainID: "osmosis-1", ChannelID: "channel-216", PortID: "transfer", Sequence: 109565}
	pp.markPacketPruned(key, prunedPacketRecord{QueryChainID: "osmosis-1", QueryEvent: "send_packet"})
	if !pp.isPacketPruned("osmosis-1", "channel-216", "transfer", 109565) {
		t.Fatal("expected packet to be blacklisted")
	}
	if pp.isPacketPruned("osmosis-1", "channel-216", "transfer", 109564) {
		t.Fatal("expected different sequence to not be blacklisted")
	}
}
