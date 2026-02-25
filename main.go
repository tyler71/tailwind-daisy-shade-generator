package main

import (
    "flag"
    "fmt"
    "math"
    "os"
    "regexp"
    "strconv"
    "strings"
    "path/filepath"
)

type OKLCH struct {
    L float64 // 0..1
    C float64 // typically 0..0.322
    H float64 // degrees 0..360
}

type OKLab struct {
    L float64
    A float64
    B float64
}

func main() {
    name := flag.String("name", "primary", "CSS variable name prefix (e.g. 'primary' -> --color-primary-500)")
    flag.Parse()

    appName := filepath.Base(os.Args[0]);

    if flag.NArg() < 1 {
        fmt.Println("Usage: " + appName + " [--name primary] <color>")
        fmt.Println("Examples:")
        fmt.Printf("  %s --name brand \"oklch(63.092%% 0.20837 34.93)\"\n", appName)
        fmt.Println("  " + appName + " \"#0ea5e9\"")
        fmt.Println("  " + appName + " \"hsl(200 80% 50%)\"")
        os.Exit(1)
    }

    input := strings.TrimSpace(strings.Join(flag.Args(), " "))

    baseColor, err := parseColor(input)
    if err != nil {
        fmt.Printf("Error parsing color: %v\n", err)
        os.Exit(1)
    }

    generateShades(*name, baseColor)
}

func parseColor(input string) (OKLCH, error) {
    input = strings.TrimSpace(strings.ToLower(input))

    // OKLCH: oklch(63.092% 0.20837 34.93)
    reOKLCH := regexp.MustCompile(`oklch\(\s*([\d.]+)%?\s+([\d.]+)\s+([\d.]+)\s*\)`)
    if m := reOKLCH.FindStringSubmatch(input); m != nil {
        l, _ := strconv.ParseFloat(m[1], 64)
        if strings.Contains(m[0], "%") {
            l = l / 100.0
        } else if l > 1 {
            l = l / 100.0
        }
        c, _ := strconv.ParseFloat(m[2], 64)
        h, _ := strconv.ParseFloat(m[3], 64)
        return OKLCH{L: clamp01(l), C: c, H: normalizeHue(h)}, nil
    }

    // HEX: #rgb or #rrggbb
    if strings.HasPrefix(input, "#") {
        r, g, b, err := parseHexColor(input)
        if err != nil {
            return OKLCH{}, err
        }
        return rgbToOklch(r, g, b), nil
    }

    // HSL: hsl(h s% l%) or hsl(h, s%, l%)
    reHSL := regexp.MustCompile(`hsl\(\s*([\d.]+)\s*,?\s*([\d.]+)%\s*,?\s*([\d.]+)%\s*\)`)
    if m := reHSL.FindStringSubmatch(input); m != nil {
        h, _ := strconv.ParseFloat(m[1], 64)
        s, _ := strconv.ParseFloat(m[2], 64)
        l, _ := strconv.ParseFloat(m[3], 64)
        r, g, b := hslToRgb(h, s/100.0, l/100.0)
        return rgbToOklch(r, g, b), nil
    }

    return OKLCH{}, fmt.Errorf("unsupported color format. Use oklch(), hex (#rrggbb), or hsl().")
}

func generateShades(name string, base OKLCH) {
    // Reference distribution derived from provided example; we scale relative to 500 shade
    ref := []struct {
        key string
        L   float64 // absolute L in 0..1 for reference base color
        C   float64 // absolute C for reference base color
    }{
        {"50", 0.88849, 0.05112},
        {"100", 0.85511, 0.06922},
        {"200", 0.78879, 0.10467},
        {"300", 0.72831, 0.14216},
        {"400", 0.67382, 0.17859},
        {"500", 0.63092, 0.20837},
        {"600", 0.53368, 0.17929},
        {"700", 0.42528, 0.13871},
        {"800", 0.31055, 0.09660},
        {"900", 0.18452, 0.04614},
        {"950", 0.10358, 0.02252},
    }

    const refBaseL = 0.63092
    const refBaseC = 0.20837

    // Print base without suffix
    fmt.Printf("--color-%s: oklch(%.3f%% %.5f %.3f);\n", name, base.L*100, base.C, base.H)

    for _, s := range ref {
        // scale L and C proportionally to the provided base color
        l := base.L * (s.L / refBaseL)
        l = clamp01(l)
        c := base.C * (s.C / refBaseC)
        if c < 0 { c = 0 }
        h := base.H
        fmt.Printf("--color-%s-%s: oklch(%.3f%% %.5f %.3f);\n", name, s.key, l*100, c, h)
    }
}

// ---------- Parsing helpers ----------

func parseHexColor(s string) (float64, float64, float64, error) {
    s = strings.TrimPrefix(strings.TrimSpace(s), "#")
    var r, g, b uint64
    var err error
    if len(s) == 3 {
        r, err = strconv.ParseUint(strings.Repeat(string(s[0]), 2), 16, 8)
        if err != nil { return 0, 0, 0, err }
        g, err = strconv.ParseUint(strings.Repeat(string(s[1]), 2), 16, 8)
        if err != nil { return 0, 0, 0, err }
        b, err = strconv.ParseUint(strings.Repeat(string(s[2]), 2), 16, 8)
        if err != nil { return 0, 0, 0, err }
    } else if len(s) == 6 {
        r, err = strconv.ParseUint(s[0:2], 16, 8)
        if err != nil { return 0, 0, 0, err }
        g, err = strconv.ParseUint(s[2:4], 16, 8)
        if err != nil { return 0, 0, 0, err }
        b, err = strconv.ParseUint(s[4:6], 16, 8)
        if err != nil { return 0, 0, 0, err }
    } else {
        return 0, 0, 0, fmt.Errorf("invalid hex color length")
    }
    return float64(r)/255.0, float64(g)/255.0, float64(b)/255.0, nil
}

func hslToRgb(h, s, l float64) (float64, float64, float64) {
    h = math.Mod(h, 360.0)
    if h < 0 { h += 360 }

    c := (1 - math.Abs(2*l-1)) * s
    x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
    m := l - c/2

    var r1, g1, b1 float64
    switch {
    case h < 60:
        r1, g1, b1 = c, x, 0
    case h < 120:
        r1, g1, b1 = x, c, 0
    case h < 180:
        r1, g1, b1 = 0, c, x
    case h < 240:
        r1, g1, b1 = 0, x, c
    case h < 300:
        r1, g1, b1 = x, 0, c
    default:
        r1, g1, b1 = c, 0, x
    }
    return r1 + m, g1 + m, b1 + m
}

// ---------- Color space conversions ----------

func rgbToOklch(r, g, b float64) OKLCH {
    lab := rgbToOklab(r, g, b)
    c := math.Hypot(lab.A, lab.B)
    h := math.Atan2(lab.B, lab.A) * 180 / math.Pi
    if h < 0 { h += 360 }
    return OKLCH{L: lab.L, C: c, H: h}
}

func rgbToOklab(r, g, b float64) OKLab {
    // to linear sRGB
    rl := srgbToLinear(r)
    gl := srgbToLinear(g)
    bl := srgbToLinear(b)

    // Matrix to LMS (from Björn Ottosson's OKLab spec)
    l := 0.4122214708*rl + 0.5363325363*gl + 0.0514459929*bl
    m := 0.2119034982*rl + 0.6806995451*gl + 0.1073969566*bl
    s := 0.0883024619*rl + 0.2817188376*gl + 0.6299787005*bl

    l_ := cbrt(l)
    m_ := cbrt(m)
    s_ := cbrt(s)

    L := 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_
    A := 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_
    B := 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_

    return OKLab{L: L, A: A, B: B}
}

func srgbToLinear(c float64) float64 {
    if c <= 0.04045 {
        return c / 12.92
    }
    return math.Pow((c+0.055)/1.055, 2.4)
}

// ---------- utils ----------

func cbrt(x float64) float64 { // real cube root
    if x < 0 {
        return -math.Pow(-x, 1.0/3.0)
    }
    return math.Pow(x, 1.0/3.0)
}

func clamp01(x float64) float64 {
    if x < 0 { return 0 }
    if x > 1 { return 1 }
    return x
}

func normalizeHue(h float64) float64 {
    h = math.Mod(h, 360.0)
    if h < 0 { h += 360 }
    return h
}
