# Assets

PNGs with an SVG beside them are rendered from it (`rsvg-convert -w <size> x.svg -o x.png`).

| File | Used for | Size |
| --- | --- | --- |
| `icon.png` (`icon.svg`) | The app icon: macOS Dock, Windows executable, installer and window, Linux menu, the website. | 1024×1024 |
| `icon-ios.png` | iOS home screen: opaque and full-bleed, iOS rounds the corners itself. | 1024×1024 |
| `android-foreground.png`, `android-background.png`, `android-monochrome.png` | Android's adaptive icon layers; the artwork stays inside the middle 66dp of the 108dp canvas. | 432×432 |
| `logo.png` | The icon inside the app. | 192×192 |
| `tray.png`, `tray-template.png` (`tray-template.svg`) | The tray icon in color (Windows, Linux) and the template macOS tints. | 64×64, 44×44 |
| `fonts/` | Inter, the UI's typeface (SIL Open Font License, `fonts/LICENSE.txt`). | |
