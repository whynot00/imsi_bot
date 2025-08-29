package models

const (
	Date           = 2
	Standart       = 4
	OpertorCode    = 7
	Coords         = 11
	SignalStrength = 12
	IMSI           = 14
	IMEI           = 15
)

var IMSIOperators = map[string]string{
	"25001": "MTS",
	"25002": "MegaFon",
	"25003": "Rostelecom",
	"25005": "ETK",
	"25006": "Danycom",
	"25007": "Smarts",
	"25008": "Vainah Telecom",
	"25009": "Skylink / Sotel",
	"25010": "Dontel (MTS)",
	"25011": "Yota",
	"25012": "Baikalwestcom",
	"25013": "Kuban GSM",
	"25014": "MegaFon",
	"25016": "Miatel",
	"25017": "Utel (не используется)",
	"25018": "Astran (MVNO Tele2)",
	"25019": "Alfa-Mobile (MVNO Beeline)",
	"25020": "Tele2",
	"25023": "GTNT",
	"25026": "VTB Mobile (MVNO Tele2)",
	"25027": "Letai (Tattelecom)",
	"25030": "Ostelecom",
	"25032": "K-Telecom (WIN Mobile)",
	"25033": "Sevmobile",
	"25034": "Krymtelecom",
	"25035": "Motiv",
	"25037": "MCN Telecom",
	"25039": "Rostelecom (AKOS и др.)",
	"25040": "Voentelecom (MVNO Tele2)",
	"25042": "MTT",
	"25043": "Sprint",
	"25045": "Gazprombank Mobile (MVNO Tele2)",
	"25047": "Next Mobile / GorodMobile",
	"25048": "V-Tell",
	"25050": "SberMobile (MVNO Tele2)",
	"25051": "Center 2M",
	"25054": "Miranda-Media",
	"25055": "NP GLONASS",
	"25059": "Wifire (MVNO MegaFon)",
	"25060": "Volna Mobile (Crimea, MVNO)",
	"25062": "T-Mobile (ex Tinkoff Mobile)",
	"25077": "GLONASS (повторный)",
	"25091": "MegaFon (Sonic Duo)",
	"25092": "MTS (Primtelefon)",
	"25094": "MirTelecom",
	"25096": "7Telecom (K-Telecom)",
	"25097": "Phoenix (DNR)",
	"25098": "Lugacom (LNR)",
	"25099": "Beeline / Extel / NTK",
}
