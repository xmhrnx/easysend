# EasySend privacy policy

_Effective 9 November 2026_

EasySend sends files, folders and text between your devices and to other people. It is
built to keep what you send between you and the people you send it to. It has no
accounts, no advertising, no analytics and no tracking, and the developer never receives
the files or text you transfer.

This policy explains what EasySend does with data, on every platform it runs on
(Windows, macOS, Linux, Android and iOS).

## What stays on your device

- **Your settings and transfer history**: device name, preferences, favorite devices and the list of past
  transfers are stored only on your device, in EasySend's own data folder.
- **Files you receive** are saved on your device: in your Downloads folder (EasySend's
  folder in the app's Documents on iPhone and iPad), or a folder you choose.
- **Crash information**: if EasySend crashes, a description of the crash is saved on
  your device. It is sent only if you choose to attach it to a problem report.
- **A short networking log** (which network adapters and ports EasySend uses, which
  devices it found) is kept in memory while the app runs and is sent only if you choose
  to attach it to a problem report.

## Sending to nearby devices

When EasySend is visible to nearby devices, it announces itself on your local network,
as the LocalSend protocol does: your device name, device type and model, an identifier
for your device, and the port it listens on. Other devices on the same network running
EasySend or LocalSend can see this. You can turn this off in Settings → Network →
Visible to nearby devices.

Files and text go directly from one device to the other over your local network,
encrypted with HTTPS. They don't pass through any server.

## Sending over the internet with a code

Internet transfers use the Magic Wormhole protocol. The two devices meet through a
rendezvous server and, when they can't connect directly, a relay server. By default
these are the public servers run by the Magic Wormhole project
(`relay.magic-wormhole.io` and `transit.magic-wormhole.io`). You can choose your own in
Settings → Internet transfers.

Everything you send is end-to-end encrypted with a key derived from the code, so these
servers can't read it. Like any server, they can see your IP address, when you connect
and how much data is transferred.

## Sharing with a link or QR code

A link is served by EasySend on your own device. Anyone who has the link (and the PIN,
if you set one) can open it and download what you shared, until you stop sharing or the
link expires. These pages use plain HTTP, so on a shared network others could observe
the transfer: use a PIN and an expiry for anything sensitive, or an internet code.

When "Receive from a browser" is on, anyone with its link can upload files and text to
your device.

When "Open browser links to the internet" is on, EasySend asks your router (with UPnP)
to forward a port to your device so links work from outside your network. Your router
then accepts connections on that port from the internet.

## Reporting a problem

If you use Report a problem, what you write is sent to a relay server run by the
developer on Cloudflare, which posts it as an issue in EasySend's **public** GitHub
repository (github.com/xmhrnx/easysend). A report contains:

- the summary and details you type;
- if you leave them ticked: the app version and build, your operating system and its
  version, processor type, device model and language;
- if you tick them: the last crash, and the networking log (network adapter names, the
  local IP addresses of your device, ports, and the names and local addresses of
  devices EasySend found or that connected to it).

The form shows what will be sent before you send it. Don't include anything you don't
want to be public.

If you leave an email address so the developer can reply, it is **not** published: it
is stored privately on the relay for one year, then deleted. To limit abuse, the relay
counts reports from each IP address; these counts are deleted after about a day.

Cloudflare's and GitHub's privacy policies also apply to reports.

## Donations and tips

On Windows, Linux and Macs not installed from the App Store, the Donate button opens the
developer's page on Ko-fi in your browser; Ko-fi's privacy policy applies there.

In apps installed from the App Store or Google Play, tips are in-app purchases processed
by Apple or Google. The developer receives no payment details; Apple's and Google's
privacy policies apply. Tips unlock nothing in the app.

## Permissions

EasySend asks only for what it uses:

- **Local network** access, to find and exchange files with nearby devices.
- **Notifications**, to tell you when something arrives (optional).
- **Files**: only the files you choose to send, and the folder received files go to.
- On desktops, **start at login**, only if you turn it on.

## Children

EasySend doesn't collect personal information from anyone, including children.

## Changes

If this policy changes, the new version will be published here with a new effective
date.

## Contact

Questions about privacy: use Report a problem in the app, or open an issue at
https://github.com/xmhrnx/easysend/issues.
