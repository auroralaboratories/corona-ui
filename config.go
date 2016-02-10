package main

type Color struct {
    Red   float64     `yaml:"red"`
    Green float64     `yaml:"green"`
    Blue  float64     `yaml:"blue"`
    Alpha float64     `yaml:"alpha"`
}

type WindowConfig struct {
    Width       int     `yaml:"width"`
    Height      int     `yaml:"height"`
    X           int     `yaml:"x"`
    Y           int     `yaml:"y"`
    Background  Color   `yaml:"background"`
    Frame       bool    `yaml:"frame"`
    Position    string  `yaml:"position"`
    Resizable   bool    `yaml:"resizable"`
    Stacking    string  `yaml:"stacking"`
    Transparent bool    `yaml:"transparent"`
    Type        string  `yaml:"type"`
}

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