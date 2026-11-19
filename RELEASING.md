# Releasing EasySend

Publishing a GitHub release with a tag like `v1.0.0` runs
`.github/workflows/release.yml`. The tag sets the version (a four-part tag such
as `v1.4.2.7` also sets the build number; otherwise the run number does, so
every store upload gets a new one). The workflow builds every package and
attaches them all to the release.

The workflow is written by the Giode CLI from the `[ci]` section of
`giode.toml`: run `giode ci -f` instead of editing it, then redo the two pins by
hand (the CLI at `giode-cli/cmd/giode@v0.1.35-alpha.1` instead of `@latest`,
and `go-version-file: go.mod` instead of `go-version: stable`). If a platform
fails and the release gets no files, Actions → Attach builds attaches what the
run built.

| Where | What the release gets | Needs |
| --- | --- | --- |
| GitHub, Linux | `.deb` and `.rpm`, x64 | nothing |
| GitHub, Windows | a `setup.exe` installer, x64 | nothing; a code signing certificate stops SmartScreen warnings |
| GitHub, Android | an `.apk` | the release keystore, or it is signed with a debug key |
| GitHub, macOS | a universal `.dmg` | a Developer ID certificate, or macOS refuses to open it |
| Microsoft Store | an `.msix` to upload in Partner Center | the Partner Center account (already set up) |
| Mac App Store | a `.pkg` to upload with Transporter | an Apple Developer account and the store certificates |
| Google Play | an `.aab` to upload in the Play Console | a Play developer account and the release keystore |
| App Store (iOS) | an `.ipa` to upload with Transporter | the iOS certificate and both provisioning profiles |

## Repository secrets

Set these in GitHub under Settings → Secrets and variables → Actions. Each
`*_BASE64` secret is the file encoded with `base64 -i file | pbcopy` on macOS,
or `base64 -w0 file` on Linux. Without a secret, that step is skipped (or the
package is left unsigned) and the rest of the release still builds.

| Secret | What it is |
| --- | --- |
| `ANDROID_KEYSTORE_BASE64`, `GIODE_KEYSTORE_PASS` | The Android release keystore (`release.keystore`) and its password: signs the `.apk` and the `.aab` (your Play upload key) |
| `WINDOWS_CERT_BASE64`, `GIODE_WINDOWS_CERT_PASS` | An Authenticode `.pfx` and its password: signs the Windows installer (optional) |
| `MACOS_CERT_BASE64`, `MACOS_CERT_PASS` | The "Developer ID Application" certificate as a `.p12`: signs the `.dmg` |
| `GIODE_NOTARY_ID`, `GIODE_NOTARY_TEAM_ID`, `GIODE_NOTARY_PASS` | Your Apple ID, team ID and an app-specific password: notarizes the `.dmg` |
| `MACOS_STORE_CERT_BASE64`, `MACOS_STORE_CERT_PASS` | The "Apple Distribution" certificate as a `.p12` |
| `MACOS_INSTALLER_CERT_BASE64`, `MACOS_INSTALLER_CERT_PASS` | The "Mac Installer Distribution" certificate as a `.p12` |
| `MACOS_PROFILE_BASE64` | The Mac App Store provisioning profile (`.provisionprofile`) |
| `IOS_CERT_BASE64`, `IOS_CERT_PASS` | The "Apple Distribution" certificate as a `.p12` (the same one as for the Mac works) |
| `IOS_PROFILE_BASE64` | The App Store provisioning profile for `app.easysend` (`.mobileprovision`) |
| `IOS_SHARE_PROFILE_BASE64` | The App Store provisioning profile for the share extension, `app.easysend.share` |
| `IOS_TEAM_ID` | Your Apple team ID |

Never commit a keystore or profile: `*.keystore` is in `.gitignore`, and the
workflow writes `release.keystore` from the secret, which `android.keystore` in
`giode.toml` points at.

## Mac App Store

1. Join the Apple Developer Program.
2. In Certificates, Identifiers & Profiles:
   - Register the App ID `app.easysend` (the `app.id` in `giode.toml`).
   - Create an "Apple Distribution" and a "Mac Installer Distribution" certificate.
   - Create a "Mac App Store" provisioning profile for the App ID.
3. In App Store Connect, create the macOS app with the same bundle ID (one
   record can hold both the iOS and the macOS app).
4. Add the three Mac App Store secrets above, then publish a release. Upload
   `EasySend-<version>.pkg` with the Transporter app.
5. In App Store Connect, fill in the listing (copy below). **Category:**
   Utilities, **export compliance** and the privacy manifest are in the build.

The sandbox build has network client and server access (nearby transfers,
links, internet codes), the files and folders the user picks, and Downloads,
where received files go by default.

## App Store (iOS)

1. In Certificates, Identifiers & Profiles:
   - Register the App IDs `app.easysend` and `app.easysend.share` (the share
     extension), and the App Group `group.app.easysend`; enable the group on
     both App IDs.
   - Create two "App Store" provisioning profiles, one for each App ID, with
     the "Apple Distribution" certificate.
2. Create the iOS app in App Store Connect with the bundle ID `app.easysend`.
3. Add the `IOS_*` secrets, then publish a release. Upload `EasySend.ipa` with
   Transporter. Without both profiles the job only checks the simulator build.
4. Create the in-app purchases the Donate button offers, as **consumables**:
   `tip_coffee`, `tip_lunch` and `tip_dinner`. They unlock nothing.
5. In App Review notes, explain that discovery needs a second device on the
   same network (or LocalSend on any device), and that internet codes work
   with any Magic Wormhole client.

EasySend finds nearby devices with multicast. iOS only allows multicast to apps
with Apple's Multicast Networking entitlement
(`com.apple.developer.networking.multicast`), which you request from Apple. Try
discovery on a device from TestFlight; if iPhones don't find or aren't found,
request the entitlement (devices can still be added by IP address).

## Google Play

1. Open a Play Console developer account and create the app, package name
   `app.easysend` (the `app.id`).
2. Create the upload keystore once: `GIODE_KEYSTORE_PASS=... giode keystore`
   writes `release.keystore`. Keep it and its password safe and out of git.
   Add it as `ANDROID_KEYSTORE_BASE64` and the password as
   `GIODE_KEYSTORE_PASS`.
3. Turn on **Play App Signing** (the default for new apps): Google signs the
   app for users, and your keystore is only the upload key.
4. Publish a release and upload `EasySend.aab` to a testing track first.
5. Play requires the latest target API: `target_sdk = 36` (Android 16) is set
   in `giode.toml`, since Google Play requires API 36 for new apps and updates
   since 31 August 2026. The build compiles against the newest Android
   platform installed on the runner.
   Targeting 36 turns on predictive back, and Gio handles back in
   `onBackPressed()`, which Android 16 then no longer calls; so
   `application_attributes` sets `enableOnBackInvokedCallback="false"`.
   Before each Play upload, test the back gesture on an Android 16 device:
   it must go back a screen (out of Settings, out of a transfer), not close
   the app.
6. Declarations in App content:
   - **Foreground service** (`FOREGROUND_SERVICE_DATA_SYNC`): transfers the user
     starts keep running while the app is in the background. Play asks for a
     short video of a transfer continuing with the app in the background.
   - **Data safety:** see below.
   - **Ads:** none. **Target audience:** 13+ (general audience, not for children).
   - **Content rating:** a utility with no objectionable content; users can
     exchange files with each other directly (no shared or public content).
7. Create the in-app products `tip_coffee`, `tip_lunch` and `tip_dinner`
   (consumable), as on iOS.

## Microsoft Store

EasySend is already in the Store (product `9NRD1RDP4J3C`).
`[windows.msix]` in `giode.toml` holds the Product identity values
(`MehranZahir.EasySend`, `CN=B40E8F57-A7E0-40BB-8E74-5D051D0B578B`,
"Mehran Zahiroddini"). For each release, upload `EasySend-<version>-amd64.msix`
in a new submission; the Store signs it. The package declares its firewall
rules (`firewall = true`) and a startup task (start at login, off until the
user turns it on). Explain `runFullTrust` as: a desktop app that serves files
to other devices on the local network.

## Privacy and store answers

**Privacy policy URL:** https://xmhrnx.github.io/easysend/#privacy (the
summary on the site, which links to the full policy,
https://github.com/xmhrnx/easysend/blob/main/PRIVACY.md; give the full policy's
URL where a store wants the complete text). Keep the two consistent.

Nothing is collected unless the user sends a problem report, which goes to the
developer's relay and becomes an issue in the public GitHub repository:

- **Apple App Privacy** (matches `[privacy]` in `giode.toml`): Data linked to
  the user, not used for tracking, purpose App Functionality: Contact Info →
  Email Address (optional); User Content → Customer Support; Diagnostics →
  Crash Data, Other Diagnostic Data.
- **Google Play Data safety:** collected, not shared for advertising, optional,
  user-initiated, encrypted in transit: Personal info → Email address; App
  activity → Other user-generated content (the report text); App info and
  performance → Crash logs, Diagnostics. Purpose: App functionality (support).
  Deletion: users can ask through an issue or a report. Tips are processed by
  Google Play; the developer receives no payment details.
- **Export compliance:** standard encryption only (HTTPS, Magic Wormhole's
  SPAKE2 and NaCl); `ITSAppUsesNonExemptEncryption = false` is in the build.
- **Age rating:** 4+ / Everyone. Users can exchange files with each other, so
  answer yes where a questionnaire asks about sharing content between users
  only if it covers direct, private transfers.

## Store listing

**Name:** EasySend

**Subtitle / short description (≤ 80 characters):**
Send files, folders and text nearby, over the internet, or to any browser.

**Keywords (Apple, ≤ 100 characters):**
file transfer,send files,share,airdrop,localsend,wormhole,wifi,nearby,qr code,receive,folder

**Description:**

> EasySend sends files, folders and text between any devices, three ways:
>
> Nearby: devices on the same network find each other and transfer directly,
> with no cloud in between. Works with LocalSend on every platform.
>
> Over the internet: create a short code like 7-crossover-clockwork and the
> other side types it in. Transfers are end-to-end encrypted, and work with any
> Magic Wormhole client.
>
> To any browser: share a link or QR code that anyone can open to download,
> with an optional expiry, download limit and PIN. Let people upload to you
> from their browser too.
>
> No accounts, no ads, no tracking. Free and open source.

**Category:** Utilities (Apple), Tools (Google Play), Utilities & tools
(Microsoft Store).

**Screenshots:** the shots in `docs/screenshots/` show the right screens but
are 1×: desktop `send.png`, `receive.png` (1180 × 800), phone
`phone-send.png`, `phone-accept.png`, `phone-settings.png` (393 × 852). Capture
them again at store sizes before submitting: Mac 2880 × 1800, iPhone 6.9"
1320 × 2868 (plus iPad 13" 2064 × 2752 if the app runs on iPad), Play phone
1080 × 1920 to 1080 × 2160 (Play needs the long side at most twice the short
one), Microsoft Store 1920 × 1080.
