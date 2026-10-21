package com.nearby.chat.feature.chat.ui

import androidx.compose.ui.graphics.Color
import com.nearby.chat.core.model.RoomPlaceType
import com.nearby.chat.core.theme.AppColors

data class RoomPlaceTypeColors(
    val fg: Color,
    val bg: Color,
) {
    companion object {
        fun of(placeType: Int): RoomPlaceTypeColors = when (RoomPlaceType.option(placeType).value) {
            RoomPlaceType.COMMUNITY -> RoomPlaceTypeColors(AppColors.BrandDark, AppColors.BrandContainer)
            RoomPlaceType.SCHOOL -> RoomPlaceTypeColors(AppColors.TealStrong, AppColors.BrandContainer)
            RoomPlaceType.BUSINESS -> RoomPlaceTypeColors(AppColors.OrangeStrong, AppColors.BrandContainer)
            RoomPlaceType.SCENIC -> RoomPlaceTypeColors(AppColors.SkyStrong, AppColors.BrandContainer)
            RoomPlaceType.CAMPUS -> RoomPlaceTypeColors(AppColors.LimeStrong, AppColors.BrandContainer)
            RoomPlaceType.TRANSIT -> RoomPlaceTypeColors(AppColors.NeutralStrong, AppColors.NeutralContainer)
            else -> RoomPlaceTypeColors(AppColors.NeutralMedium, AppColors.NeutralContainerLight)
        }
    }
}
