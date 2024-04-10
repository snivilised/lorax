package enums

type ItemDiscriminator uint32

const (
	// ItemDiscNative enum value that represents the native type T.
	//
	ItemDiscNative ItemDiscriminator = 0

	// ItemDiscBoolean enum value that represents a general boolean value
	// typically used by predicate based operations eg All.
	//
	ItemDiscBoolean ItemDiscriminator = 1 << (iota - 1)

	// ItemDiscChan enum value that represents a channel of T
	//
	ItemDiscChan

	// ItemDiscError enum value that represents an error
	//
	ItemDiscError

	// ItemDiscNumeric enum value that represents a general numeric value
	// typically used by range operations that require a number.
	//
	ItemDiscNumeric

	// ItemDiscPulse enum value that represents a Tick value.
	//
	ItemDiscPulse

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
)

var ItemDescriptions map[ItemDiscriminator]string

func init() {
	ItemDescriptions = itemsDiscDescriptions{
		ItemDiscNative:    "native",
		ItemDiscBoolean:   "boolean",
		ItemDiscChan:      "channel",
		ItemDiscError:     "error",
		ItemDiscNumeric:   "numeric",
		ItemDiscPulse:     "pulse",
		ItemDiscTickValue: "tick",
		ItemDiscOpaque:    "opaque",
	}
}
