package com.nearby.chat.core.theme

import androidx.compose.runtime.Composable
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.colorResource
import com.nearby.chat.R

object AppColors {
    val Brand = Color(0xFF4C8DFF)
    val BrandLight = Color(0xFF7EB0FF)
    val BrandDark = Color(0xFF2A6BD8)
    val BrandContainer = Color(0xFFEEEEEE)

    val PageBg: Color
        @Composable
        @ReadOnlyComposable
        get() = colorResource(R.color.page_bg)
    val CardBg = Color.White
    val Surface = Color.White

    val TextPrimary = Color(0xFF1F1F1F)
    val TextSecondary = Color(0xFF666666)
    val TextTertiary = Color(0xFF999999)
    val TextMuted = Color(0xFF888888)
    val TextDeep = Color(0xFF2A2A2A)

    val Border = Color(0x0F000000)
    val BorderSoft = Color(0xFFE8E8E8)
    val Divider = Color(0x0F000000)

    val Danger = Color(0xFFE05A52)
    val Success = Color(0xFF36B37E)
    val Online = Color(0xFF36B37E)
    val Warning = Color(0xFFE09A45)
    val Reward = Color(0xFF36B37E)

    val SheetHandle = Color(0x1F000000)

    val ComposerBg = Color.White

    val ComposerIcon = TextPrimary

    val SwitchOff: Color
        @Composable
        @ReadOnlyComposable
        get() = PageBg

    val SendGradient = Brush.linearGradient(
        colors = listOf(Brand, Brand),
    )

    val MediaPlaceholder = Color(0xFFF2F2F2)
    val MediaScrim = Color.Black.copy(alpha = 0.45f)
    val OverlayFaint = Color(0x08000000)
    val OverlaySoft = Color(0x0F000000)
    val OverlayHairline = Color(0x0A000000)
    val OverlayMuted = Color(0x0D000000)
    val NeutralContainer = Color(0xFFF0F0F0)
    val NeutralContainerLight = Color(0xFFF5F5F5)
    val NeutralLine = Color(0xFFE5E5E5)
    val NeutralShade = Color(0xFFD8D8D8)
    val NeutralStrong = Color(0xFF555555)
    val NeutralMedium = Color(0xFF6B6B6B)
    val ThumbPlaceholder = NeutralContainer
    val CountScrim = Color(0x94000000)
    val ScrimMedium = Color(0x66000000)
    val ScrimStrong = Color(0xB3000000)
    val UploadScrim = ScrimMedium
    val UploadFailScrim = Color(0x99000000)
    val UploadDeleteBg = ScrimStrong
    val ToastBg = ScrimStrong
    val ShadowFaint = Color(0x10000000)
    val ShadowSoft = Color(0x33000000)
    val BorderStrong = Color(0x22000000)
    val BrandDarkBorder = Color(0x472A6BD8)

    val MenuDarkBg = Color(0xFF2C2C2C)
    val MenuShadow = ScrimMedium

    val RecordCancel = Color(0xFFE54D42)
    val RecordCancelTint = Color(0x1AE54D42)

    val ReviewPending = Color(0xFFC9952A)
    val ReviewApproved = Color(0xFF2F9B63)
    val ReviewRejected = Color(0xFFD14343)

    val AdBadgeText = Color(0xFF909090)
    val AdBadgeBorder = Color(0xFFD0D0D0)

    val TealStrong = Color(0xFF0F766E)
    val SkyStrong = Color(0xFF0369A1)
    val LimeStrong = Color(0xFF4D7C0F)

    val BrandTint = Color(0x1F4C8DFF)
    val SuccessTint = Color(0x1F10B981)
    val SuccessStrong = Color(0xFF047857)
    val WarningTint = Color(0x1FF59E0B)
    val WarningTintLight = Color(0x1AF59E0B)
    val WarningStrong = Color(0xFFB45309)
    val DangerTint = Color(0x1EE05A52)
    val DangerStrong = Color(0xFFC24740)
    val DangerDeep = Color(0xFFB91C1C)
    val OrangeStrong = Color(0xFFC2410C)
    val OrangeTint = Color(0x24FF976A)
    val NeutralTint = Color(0x14000000)

    val TimelineDot = Color(0xFFD9D9D9)
    val TimelineLine = Color(0xFFECECEC)
}
