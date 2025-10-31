# Hexadecimal Line Numbers Reference

Files are numbered in hexadecimal to save screen space. This is especially useful when you have 100+ endpoints.

## Quick Reference

| Hex | Decimal | Notes            |
| --- | ------- | ---------------- |
| 1   | 1       | First file       |
| A   | 10      |                  |
| F   | 15      |                  |
| 10  | 16      |                  |
| 14  | 20      |                  |
| 19  | 25      |                  |
| 1E  | 30      |                  |
| 32  | 50      |                  |
| 64  | 100     | Common milestone |
| C8  | 200     |                  |
| FF  | 255     |                  |
| 100 | 256     |                  |
| 12C | 300     |                  |
| 190 | 400     |                  |
| 1F4 | 500     |                  |
| 3E8 | 1000    |                  |

## Space Savings

With 1000 endpoints:

- **Decimal**: "1000" = 4 characters per line number
- **Hex**: "3E8" = 3 characters per line number
- **Saved**: 1 character per line = more room for filenames!

## How to Use

### Goto Command

Press `:` then type the hex number you see in the sidebar:

- See file numbered `64`? Type `:64` to jump there
- See file numbered `FF`? Type `:FF` to jump there
- See file numbered `3E8`? Type `:3E8` to jump there

### Converting in Your Head

**Common patterns:**

- `0-9` = same as decimal (0-9)
- `A` = 10
- `B` = 11
- `C` = 12
- `D` = 13
- `E` = 14
- `F` = 15

**For larger numbers:**

- First digit × 16, plus second digit
- Examples:
  - `1A` = (1 × 16) + 10 = 26
  - `2F` = (2 × 16) + 15 = 47
  - `64` = (6 × 16) + 4 = 100

**Pro tip:** You don't need to convert! Just use the hex numbers directly as shown in the sidebar.

## Why Hexadecimal?

1. **Compact**: Represents larger numbers in fewer characters
2. **Familiar**: Used in programming, memory addresses, color codes (#FF0000)
3. **Screen real estate**: More space for long filenames in nested directories

## Quick Conversion Tools

If you need to convert:

```bash
# Decimal to hex
printf '%X\n' 100  # Output: 64

# Hex to decimal
echo $((16#64))    # Output: 100
```

Or just use the search feature (`Ctrl+R`) and type the filename instead!
