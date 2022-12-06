package utils

type DivineInfo struct {
	Title           string
	ImagePath       string
	InscriptionPath string
}

var Divinations = [...]*DivineInfo{
	{
		Title:           "【愚者】- 正位",
		ImagePath:       "tarot/card/0_fool.jpg",
		InscriptionPath: "tarot/inscription/0_fool.txt",
	},
	{
		Title:           "【愚者】- 逆位",
		ImagePath:       "tarot/card/0_fool_rev.jpg",
		InscriptionPath: "tarot/inscription/0_fool_rev.txt",
	},
	{
		Title:           "【魔术师】- 正位",
		ImagePath:       "tarot/card/1_magician.jpg",
		InscriptionPath: "tarot/inscription/1_magician.txt",
	},
	{
		Title:           "【魔术师】- 逆位",
		ImagePath:       "tarot/card/1_magician_rev.jpg",
		InscriptionPath: "tarot/inscription/1_magician_rev.txt",
	},
	{
		Title:           "【女祭司】- 正位",
		ImagePath:       "tarot/card/2_high_priestess.jpg",
		InscriptionPath: "tarot/inscription/2_high_priestess.txt",
	},
	{
		Title:           "【女祭司】- 逆位",
		ImagePath:       "tarot/card/2_high_priestess_rev.jpg",
		InscriptionPath: "tarot/inscription/2_high_priestess_rev.txt",
	},
	{
		Title:           "【女皇】- 正位",
		ImagePath:       "tarot/card/3_empress.jpg",
		InscriptionPath: "tarot/inscription/3_empress.txt",
	},
	{
		Title:           "【女皇】- 逆位",
		ImagePath:       "tarot/card/3_empress_rev.jpg",
		InscriptionPath: "tarot/inscription/3_empress_rev.txt",
	},
	{
		Title:           "【帝王】- 正位",
		ImagePath:       "tarot/card/4_emperor.jpg",
		InscriptionPath: "tarot/inscription/4_emperor.txt",
	},
	{
		Title:           "【帝王】- 逆位",
		ImagePath:       "tarot/card/4_emperor_rev.jpg",
		InscriptionPath: "tarot/inscription/4_emperor_rev.txt",
	},
	{
		Title:           "【教皇】- 正位",
		ImagePath:       "tarot/card/5_hierophant.jpg",
		InscriptionPath: "tarot/inscription/5_hierophant.txt",
	},
	{
		Title:           "【教皇】- 逆位",
		ImagePath:       "tarot/card/5_hierophant_rev.jpg",
		InscriptionPath: "tarot/inscription/5_hierophant_rev.txt",
	},
	{
		Title:           "【爱侣】- 正位",
		ImagePath:       "tarot/card/6_lovers.jpg",
		InscriptionPath: "tarot/inscription/6_lovers.txt",
	},
	{
		Title:           "【爱侣】- 逆位",
		ImagePath:       "tarot/card/6_lovers_rev.jpg",
		InscriptionPath: "tarot/inscription/6_lovers_rev.txt",
	},
	{
		Title:           "【战车】- 正位",
		ImagePath:       "tarot/card/7_chariot.jpg",
		InscriptionPath: "tarot/inscription/7_chariot.txt",
	},
	{
		Title:           "【战车】- 逆位",
		ImagePath:       "tarot/card/7_chariot_rev.jpg",
		InscriptionPath: "tarot/inscription/7_chariot_rev.txt",
	},
	{
		Title:           "【坚力】- 正位",
		ImagePath:       "tarot/card/8_strength.jpg",
		InscriptionPath: "tarot/inscription/8_strength.txt",
	},
	{
		Title:           "【坚力】- 逆位",
		ImagePath:       "tarot/card/8_strength_rev.jpg",
		InscriptionPath: "tarot/inscription/8_strength_rev.txt",
	},
	{
		Title:           "【隐士】- 正位",
		ImagePath:       "tarot/card/9_hermit.jpg",
		InscriptionPath: "tarot/inscription/9_hermit.txt",
	},
	{
		Title:           "【隐士】- 逆位",
		ImagePath:       "tarot/card/9_hermit_rev.jpg",
		InscriptionPath: "tarot/inscription/9_hermit_rev.txt",
	},
	{
		Title:           "【命运之轮】- 正位",
		ImagePath:       "tarot/card/10_fortune_wheel.jpg",
		InscriptionPath: "tarot/inscription/10_fortune_wheel.txt",
	},
	{
		Title:           "【命运之轮】- 逆位",
		ImagePath:       "tarot/card/10_fortune_wheel_rev.jpg",
		InscriptionPath: "tarot/inscription/10_fortune_wheel_rev.txt",
	},
	{
		Title:           "【正义】- 正位",
		ImagePath:       "tarot/card/11_justice.jpg",
		InscriptionPath: "tarot/inscription/11_justice.txt",
	},
	{
		Title:           "【正义】- 逆位",
		ImagePath:       "tarot/card/11_justice_rev.jpg",
		InscriptionPath: "tarot/inscription/11_justice_rev.txt",
	},
	{
		Title:           "【悬人】- 正位",
		ImagePath:       "tarot/card/12_hanged_man.jpg",
		InscriptionPath: "tarot/inscription/12_hanged_man.txt",
	},
	{
		Title:           "【悬人】- 逆位",
		ImagePath:       "tarot/card/12_hanged_man_rev.jpg",
		InscriptionPath: "tarot/inscription/12_hanged_man_rev.txt",
	},
	{
		Title:           "【死亡】- 正位",
		ImagePath:       "tarot/card/13_death.jpg",
		InscriptionPath: "tarot/inscription/13_death.txt",
	},
	{
		Title:           "【死亡】- 逆位",
		ImagePath:       "tarot/card/13_death_rev.jpg",
		InscriptionPath: "tarot/inscription/13_death_rev.txt",
	},
	{
		Title:           "【节制】- 正位",
		ImagePath:       "tarot/card/14_temperance.jpg",
		InscriptionPath: "tarot/inscription/14_temperance.txt",
	},
	{
		Title:           "【节制】- 逆位",
		ImagePath:       "tarot/card/14_temperance_rev.jpg",
		InscriptionPath: "tarot/inscription/14_temperance_rev.txt",
	},
	{
		Title:           "【恶魔】- 正位",
		ImagePath:       "tarot/card/15_devil.jpg",
		InscriptionPath: "tarot/inscription/15_devil.txt",
	},
	{
		Title:           "【恶魔】- 逆位",
		ImagePath:       "tarot/card/15_devil_rev.jpg",
		InscriptionPath: "tarot/inscription/15_devil_rev.txt",
	},
	{
		Title:           "【高塔】- 正位",
		ImagePath:       "tarot/card/16_tower.jpg",
		InscriptionPath: "tarot/inscription/16_tower.txt",
	},
	{
		Title:           "【高塔】- 逆位",
		ImagePath:       "tarot/card/16_tower_rev.jpg",
		InscriptionPath: "tarot/inscription/16_tower_rev.txt",
	},
	{
		Title:           "【星辰】- 正位",
		ImagePath:       "tarot/card/17_star.jpg",
		InscriptionPath: "tarot/inscription/17_star.txt",
	},
	{
		Title:           "【星辰】- 逆位",
		ImagePath:       "tarot/card/17_star_rev.jpg",
		InscriptionPath: "tarot/inscription/17_star_rev.txt",
	},
	{
		Title:           "【明月】- 正位",
		ImagePath:       "tarot/card/18_moon.jpg",
		InscriptionPath: "tarot/inscription/18_moon.txt",
	},
	{
		Title:           "【明月】- 逆位",
		ImagePath:       "tarot/card/18_moon_rev.jpg",
		InscriptionPath: "tarot/inscription/18_moon_rev.txt",
	},
	{
		Title:           "【骄阳】- 正位",
		ImagePath:       "tarot/card/19_sun.jpg",
		InscriptionPath: "tarot/inscription/19_sun.txt",
	},
	{
		Title:           "【骄阳】- 逆位",
		ImagePath:       "tarot/card/19_sun_rev.jpg",
		InscriptionPath: "tarot/inscription/19_sun_rev.txt",
	},
	{
		Title:           "【审判】- 正位",
		ImagePath:       "tarot/card/20_judgement.jpg",
		InscriptionPath: "tarot/inscription/20_judgement.txt",
	},
	{
		Title:           "【审判】- 逆位",
		ImagePath:       "tarot/card/20_judgement_rev.jpg",
		InscriptionPath: "tarot/inscription/20_judgement_rev.txt",
	},
	{
		Title:           "【世界】- 正位",
		ImagePath:       "tarot/card/21_world.jpg",
		InscriptionPath: "tarot/inscription/21_world.txt",
	},
	{
		Title:           "【世界】- 逆位",
		ImagePath:       "tarot/card/21_world_rev.jpg",
		InscriptionPath: "tarot/inscription/21_world_rev.txt",
	},
	{
		Title:           "【星币一】- 正位",
		ImagePath:       "tarot/card/22_ace_coins.jpg",
		InscriptionPath: "tarot/inscription/22_ace_coins.txt",
	},
	{
		Title:           "【星币一】- 逆位",
		ImagePath:       "tarot/card/22_ace_coins_rev.jpg",
		InscriptionPath: "tarot/inscription/22_ace_coins_rev.txt",
	},
	{
		Title:           "【圣杯一】- 正位",
		ImagePath:       "tarot/card/23_ace_cups.jpg",
		InscriptionPath: "tarot/inscription/23_ace_cups.txt",
	},
	{
		Title:           "【圣杯一】- 逆位",
		ImagePath:       "tarot/card/23_ace_cups_rev.jpg",
		InscriptionPath: "tarot/inscription/23_ace_cups_rev.txt",
	},
	{
		Title:           "【宝剑一】- 正位",
		ImagePath:       "tarot/card/24_ace_swords.jpg",
		InscriptionPath: "tarot/inscription/24_ace_swords.txt",
	},
	{
		Title:           "【宝剑一】- 逆位",
		ImagePath:       "tarot/card/24_ace_swords_rev.jpg",
		InscriptionPath: "tarot/inscription/24_ace_swords_rev.txt",
	},
	{
		Title:           "【权杖一】- 正位",
		ImagePath:       "tarot/card/25_ace_wands.jpg",
		InscriptionPath: "tarot/inscription/25_ace_wands.txt",
	},
	{
		Title:           "【权杖一】- 逆位",
		ImagePath:       "tarot/card/25_ace_wands_rev.jpg",
		InscriptionPath: "tarot/inscription/25_ace_wands_rev.txt",
	},
}
