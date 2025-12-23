# Gemini 2.5 Flash Image - Supported Aspect Ratios

**Source:** [Gemini 2.5 Flash Image Documentation](https://developers.googleblog.com/en/gemini-2-5-flash-image-now-ready-for-production-with-new-aspect-ratios/)

## Supported Aspect Ratios (10 total)

| Aspect Ratio | Native Resolution | Description | Use Case |
|--------------|-------------------|-------------|----------|
| **21:9** | 1792×768 | Ultra-wide | Cinematic panoramas |
| **16:9** | 1344×768 | Widescreen | Game backgrounds, video |
| **4:3** | 1024×768 | Classic TV | Retro displays |
| **3:2** | 1216×810 | DSLR photo | Photography standard |
| **1:1** | 1024×1024 | Square | Social media, avatars |
| **2:3** | 810×1216 | Portrait photo | Phone wallpapers |
| **3:4** | 768×1024 | Portrait classic | Tablet displays |
| **9:16** | 768×1344 | Mobile portrait | Phone screens, stories |
| **5:4** | 1024×819 | Classic monitor | Old CRT displays |
| **4:5** | 819×1024 | Instagram portrait | Social media |

## Important Notes

1. **No Upscaling**: These are the MAXIMUM native resolutions from Gemini
   - All resolutions are ~1024px equivalent (same token budget)
   - External upscaling will blur/soften the AI-generated detail
   - Use native resolution for best quality

2. **Image Size Parameter**: The `size` parameter ("1K", "2K", "4K") is passed but appears to be ignored by Gemini - native resolutions are fixed

3. **For Game Backgrounds**:
   - Use **16:9 (1344×768)** for landscape scenes
   - Use **21:9 (1792×768)** for ultra-wide panoramas
   - Accept native resolution for crisp AI-generated artwork

## CLI Usage

```bash
# Valid aspect ratios
voyage ai -generate-image -aspect "16:9" -prompt "..."
voyage ai -generate-image -aspect "21:9" -prompt "..."
voyage ai -generate-image -aspect "1:1" -prompt "..."
# ... etc for all 10 ratios

# Invalid (will error)
voyage ai -generate-image -aspect "2:1" -prompt "..."  # Not supported
```

## Validation

Scripts should validate aspect ratio against this whitelist:
```bash
VALID_RATIOS=("21:9" "16:9" "4:3" "3:2" "1:1" "2:3" "3:4" "9:16" "5:4" "4:5")
```
