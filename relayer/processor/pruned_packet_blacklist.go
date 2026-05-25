package processor

import (
	"strings"
	"time"

	"go.uber.org/zap"
)

// prunedPacketKey identifies a packet by its source chain (where the commitment lives / send occurred).
type prunedPacketKey struct {
	SourceChainID string
	ChannelID     string
	PortID        string
	Sequence      uint64
}

// prunedPacketRecord describes why a packet was blacklisted and which RPC chain could not serve events.
type prunedPacketRecord struct {
	QueryChainID          string
	QueryEvent            string
	CounterpartyChainID   string
	CounterpartyChannelID string
	CounterpartyPortID    string
	MarkedAt              time.Time
}

// isUnqueryableIBCEvent reports whether an IBC event query failed because tx/block events are unavailable (e.g. pruned RPC history).
func isUnqueryableIBCEvent(err error, event string) bool {
	return err != nil && strings.Contains(err.Error(), "no ibc messages found for "+event)
}

func (pp *PathProcessor) isPacketPruned(sourceChainID, channelID, portID string, sequence uint64) bool {
	pp.prunedPacketsMu.RLock()
	defer pp.prunedPacketsMu.RUnlock()
	_, ok := pp.prunedPackets[prunedPacketKey{
		SourceChainID: sourceChainID,
		ChannelID:     channelID,
		PortID:        portID,
		Sequence:      sequence,
	}]
	return ok
}

func (pp *PathProcessor) markPacketPruned(
	key prunedPacketKey,
	record prunedPacketRecord,
) {
	pp.prunedPacketsMu.Lock()
	defer pp.prunedPacketsMu.Unlock()
	if pp.prunedPackets == nil {
		pp.prunedPackets = make(map[prunedPacketKey]prunedPacketRecord)
	}
	if _, exists := pp.prunedPackets[key]; exists {
		return
	}
	record.MarkedAt = time.Now()
	pp.prunedPackets[key] = record
}

func (pp *PathProcessor) logPrunedPacketBlacklistSummary() {
	pp.prunedPacketsMu.RLock()
	defer pp.prunedPacketsMu.RUnlock()
	if len(pp.prunedPackets) == 0 {
		return
	}
	for key, record := range pp.prunedPackets {
		pp.log.Info("pruned packet blacklist entry",
			zap.String("packet_source_chain_id", key.SourceChainID),
			zap.String("channel_id", key.ChannelID),
			zap.String("port_id", key.PortID),
			zap.Uint64("sequence", key.Sequence),
			zap.String("event_query_chain_id", record.QueryChainID),
			zap.String("query_event", record.QueryEvent),
			zap.String("counterparty_chain_id", record.CounterpartyChainID),
			zap.String("counterparty_channel_id", record.CounterpartyChannelID),
			zap.String("counterparty_port_id", record.CounterpartyPortID),
			zap.Time("marked_at", record.MarkedAt),
		)
	}
}
