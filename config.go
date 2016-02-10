package main

type Config struct {
    Window WindowConfig `yaml:"window"`
}

func GetDefaultConfig() Config {
    return Config{
        Window: WindowConfig{
            Width:        0,
            Height:       0,
            X:           -1,
            Y:           -1,
            Background:  Color{
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