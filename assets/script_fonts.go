package assets

import _ "embed"

// Noto Sans faces for scripts Inter lacks (OFL, see fonts/NOTO-LICENSE.txt).
// CJK faces are subset (GB2312 / JIS X 0208 / KS X 1001 + punctuation) and
// Regular-only.
var (
	//go:embed fonts/NotoSansGeorgian-Regular.ttf
	NotoSansGeorgianRegular []byte
	//go:embed fonts/NotoSansGeorgian-Bold.ttf
	NotoSansGeorgianBold []byte
	//go:embed fonts/NotoSansArmenian-Regular.ttf
	NotoSansArmenianRegular []byte
	//go:embed fonts/NotoSansArmenian-Bold.ttf
	NotoSansArmenianBold []byte
	//go:embed fonts/NotoSansHebrew-Regular.ttf
	NotoSansHebrewRegular []byte
	//go:embed fonts/NotoSansHebrew-Bold.ttf
	NotoSansHebrewBold []byte
	//go:embed fonts/NotoSansArabic-Regular.ttf
	NotoSansArabicRegular []byte
	//go:embed fonts/NotoSansArabic-Bold.ttf
	NotoSansArabicBold []byte
	//go:embed fonts/NotoSansDevanagari-Regular.ttf
	NotoSansDevanagariRegular []byte
	//go:embed fonts/NotoSansDevanagari-Bold.ttf
	NotoSansDevanagariBold []byte
	//go:embed fonts/NotoSansThai-Regular.ttf
	NotoSansThaiRegular []byte
	//go:embed fonts/NotoSansThai-Bold.ttf
	NotoSansThaiBold []byte
	//go:embed fonts/NotoSansSC-Regular.ttf
	NotoSansSCRegular []byte
	//go:embed fonts/NotoSansJP-Regular.ttf
	NotoSansJPRegular []byte
	//go:embed fonts/NotoSansKR-Regular.ttf
	NotoSansKRRegular []byte
)
