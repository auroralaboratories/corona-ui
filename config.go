package main

type Config struct {
	Server Server       `json:"server"`
	Window WindowConfig `json:"window"`
}

func GetDefaultConfig() Config {
	return Config{
		Window: WindowConfig{
			Width:   ``,
			Height:  ``,
			X:       ``,
			Y:       ``,
			Monitor: -1,
			Background: &Color{
				Red:   1.0,
				Green: 1.0,
				Blue:  1.0,
				Alpha: 0.0,
			},
			Frame:       true,
			Position:    ``,
			Resizable:   true,
			Stacking:    ``,
			Transparent: false,
			Type:        ``,
		},
	}
}
