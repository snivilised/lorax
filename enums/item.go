package enums

type ItemDiscriminator uint32

const (
	ItemDiscUndefined ItemDiscriminator = 0
	// ItemDiscNative enum value that represents the native type T.
	//
	ItemDiscNative ItemDiscriminator = 1 << (iota - 1)

	// ItemDiscBoolean enum value that represents a general boolean value
	// typically used by predicate based operations eg All.
	//
	ItemDiscBoolean

	// ItemDiscWChan enum value that represents a channel of T
	//
	ItemDiscWChan

	// ItemDiscError enum value that represents an error
	//
	ItemDiscError

	// ItemDiscNumber enum value that represents a general numeric value
	// typically used by range operations that require a number.
	//
	ItemDiscNumber

	// ItemDiscTick enum value that represents a Tick value.
	//
	ItemDiscTick

	// ItemDiscTick enum value that represents a TickValue value.
	//
	ItemDiscTickValue

	// ItemDiscOpaque enum value that can be used to represent anything,
	// typically a value that is not of type T or any of the other scalar
	// types already catered for.
	ItemDiscOpaque
)

type (
	itemsDiscDescriptions map[ItemDiscriminator]string

	// CloseChannelStrategy indicates a strategy on whether to close a channel.
	CloseChannelStrategy uint32
)

const (
	// LeaveChannelOpen indicates to leave the channel open after completion.
	LeaveChannelOpen CloseChannelStrategy = iota
	// enums.CloseChannel indicates to close the channel open after completion.
	CloseChannel
)

var ItemDescriptions map[ItemDiscriminator]string

func init() {
	ItemDescriptions = itemsDiscDescriptions{
		ItemDiscNative:    "native",
		ItemDiscBoolean:   "boolean",
		ItemDiscWChan:     "write channel",
		ItemDiscError:     "error",
		ItemDiscNumber:    "number",
		ItemDiscTick:      "tick",
		ItemDiscTickValue: "tick value",
		ItemDiscOpaque:    "opaque",
	}
}
