package com.nearby.chat.feature.navigation

import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.nearby.chat.feature.map.LocationPickScreen

internal fun NavGraphBuilder.mapGraph(navigator: AppNavigator) {
    composable(
        route = Routes.LOCATION_PICK,
        arguments = listOf(
            navArgument("for") {
                type = NavType.StringType
                defaultValue = ""
            },
        ),
    ) { entry ->
        val purpose = entry.arguments?.getString("for").orEmpty()
        val navController = navigator.navController
        LocationPickScreen(
            onBack = navigator::back,
            onPicked = { result ->
                val prev = navController.previousBackStackEntry?.savedStateHandle
                if (purpose == Routes.PICK_FOR_SHARE) {
                    prev?.set(NavResultKeys.SHARE_LOCATION, result.toSharePayload())
                } else {
                    prev?.set(NavResultKeys.PICK_LAT, result.lat)
                    prev?.set(NavResultKeys.PICK_LNG, result.lng)
                    prev?.set(NavResultKeys.PICK_PLACE, result.placeName)
                }
                navController.popBackStack()
            },
        )
    }
}
