# What EasySend needs from Giode

EasySend doesn't work around these: each is general platform support that belongs in
Giode (the library) or the Giode CLI. Once they exist, the EasySend side is a few lines.

## 1. Appear in the phone share sheet (share target)

Needed for "Share → EasySend" from the gallery, Files, a browser, etc.

**Android (library + CLI)**

- CLI: declare intent filters on the app's activity from giode.toml, for example:

  ```toml
  [android.share_target]
  mime_types = ["*/*"]   # ACTION_SEND and ACTION_SEND_MULTIPLE
  text = true            # text/plain shares (links, notes)
  ```

  Today `[android] manifest` and `application` only add XML outside the activity, so the
  filters can't be attached to Gio's activity.
- Library: deliver the shared intent to Go, both at launch and while running
  (`onNewIntent`), for example:

  ```go
  Config.OnShareReceived func(s giode.SharedContent)

  type SharedContent struct {
      Files   []giode.PickedFile // EXTRA_STREAM / ClipData content:// URIs, opened as streams with Name, MIME, Size
      Text    string             // EXTRA_TEXT
      Subject string             // EXTRA_SUBJECT
  }
  ```

  The `content://` URIs must be read while the permission grant is valid, so the library
  should open them (as `PickFiles` already does) rather than hand over raw URIs.

**iOS (CLI + library)**

- CLI: build, sign and embed a Share Extension target
  (`NSExtensionPointIdentifier = com.apple.share-services`, activation rules for files,
  images, URLs and text) and put the app and the extension in a shared App Group, for
  example `[ios.share_target]` with `app_group = "group.app.easysend"` and the accepted
  types.
- Library: the extension copies the shared items into the App Group container and opens
  the app, which receives them through the same `OnShareReceived` callback.

**Desktop** already works: `[[context_actions]]` ("Send with EasySend" in the file
manager) and `SingleInstanceLock.OnLaunchArgs`.

## 2. Save received files to the phone's Downloads

On Android, received files currently stay in the app's own storage, and people move them
out with the share sheet. Giode already copies files into MediaStore Downloads internally
when sharing; exposing that would let received files land in Downloads directly:

```go
err := app.SaveToDownloads(ctx, path, "EasySend")
```

(Android: MediaStore Downloads. iOS: the app's Documents folder already shows in the
Files app through `UIFileSharingEnabled`, which giode.toml sets.)

## 3. Android multicast lock (automatic nearby discovery)

Android drops incoming multicast unless the app holds a `WifiManager.MulticastLock`, so
LocalSend-style discovery is unreliable there. EasySend falls back to scanning the subnet
and to adding devices by IP, which work without it. Either:

- a library call such as `release, err := app.AcquireMulticastLock("discovery")`, with
  the CLI adding `CHANGE_WIFI_MULTICAST_STATE`; or
- confirmation that apps should do this with their own Java through the `giode/android`
  bridge, in which case EasySend will ship it as app code.

## 4. Single-instance lock on Android and iOS (bug)

`host.LockInstance` writes its lock file in `os.TempDir()`, which is `/data/local/tmp` on
Android, where apps can't write. When the file can't be created it reports a second
instance, so `Run` returns without drawing and the app shows a white screen. EasySend now
sets `SingleInstanceLock` only on desktops, but Giode should ignore the lock on Android and
iOS (the system runs one copy of an app anyway), and an error creating the lock file
shouldn't be treated as "another copy is running".

## 5. Firewall rule in the Windows installer

EasySend serves files to phones and browsers on the local network, so Windows Firewall
has to allow incoming connections to it. Today that depends on the user clicking "Allow"
in Windows' first-run prompt (and on the network being marked Private). The setup program
from `giode build windows-installer` could add the rule, e.g. in giode.toml:

```toml
[windows.installer]
firewall = true   # netsh advfirewall firewall add rule ... program=<exe> on install, removed on uninstall
```

## 6. Splash screen color (Android and iOS)

The phone launch screen shows white before the first frame; EasySend's background is
`#0A0A0B` (dark) and `#F4F4F6` (light). Neither can be set today:

- Android: the CLI's generated `Theme.GioApp` hard-codes
  `android:windowBackground = @android:color/white`; Android 12+'s system splash takes its
  background from it too. `[android] application_attributes` can't help, since the
  activity's own theme wins.
- iOS: Info.plist gets `UILaunchScreen = true`, an empty launch screen that iOS shows
  white. A color needs `UILaunchScreen` → `UIColorName` and a named color in an asset
  catalog, which the CLI doesn't produce.

Suggested giode.toml:

```toml
[app]
splash_color = "#0A0A0B"        # both platforms; optional:
splash_color_light = "#F4F4F6"  # Android values/ vs values-night/, an iOS "Any/Dark" color set
```

On Android: `windowBackground` (and `windowSplashScreenBackground` in `values-v31`) set to
the color, plus `windowSplashScreenAnimatedIcon` from the adaptive icon. On iOS: a
`LaunchBackground` color set in Assets.car and `UILaunchScreen` = `{UIColorName =
LaunchBackground}`.

## 7. Nice to have

- `TitleBar`'s close button calls `App.Close`, which always quits, so a frameless window
  ignores `CloseToTray`. EasySend handles it through `TitleBar.OnClose` (hide when
  close-to-tray is on and the window can be hidden); the bar could do this itself, like
  the system close button.
- `app.OpenFolder(path)` to open a folder in the file manager. `Share(ShareFile)` reveals
  a file inside its parent, which suits "Show in folder", but the Settings button that
  opens the downloads folder should open the folder itself.
