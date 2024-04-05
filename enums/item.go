package enums

type ItemDiscriminator uint32

const (
	// ItemDiscNative enum value that represents the native type T.
	//
	ItemDiscNative ItemDiscriminator = 0

	// ItemDiscError enum value that represents an error
	//
	ItemDiscError ItemDiscriminator = 1 << (iota - 1)

	// ItemDiscPulse enum value that represents a Tick value.
	//
	ItemDiscPulse

	// ItemDiscTick enum value that represents a TickValue value.
	//
	ItemDiscTickValue

	// ItemDiscNumeric enum value that represents a general numeric value
	// typically used by range operations that require a number.
	//
	ItemDiscNumeric

	// ItemDiscChan enum value that represents a channel of T
	//
	ItemDiscChan
)
