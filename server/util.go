package main

func colorFromRGB(r, g, b uint32) uint32 {
	return (((r << 8) + g) << 8) + b
}
