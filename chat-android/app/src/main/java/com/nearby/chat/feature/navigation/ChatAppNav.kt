package com.nearby.chat.feature.navigation

import androidx.activity.ComponentActivity
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalTextToolbar
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.rememberNavController
import com.nearby.chat.core.share.ShareLinks
import com.nearby.chat.feature.main.MainViewModel
import com.nearby.chat.feature.update.AppUpdateHost
import com.nearby.chat.feature.update.AppUpdateViewModel

@Composable
fun ChatAppNav(mainViewModel: MainViewModel) {
    val hasToken by mainViewModel.hasToken.collectAsStateWithLifecycle(initialValue = null)
    val deepLink by mainViewModel.deepLink.collectAsStateWithLifecycle()

    val activity = LocalContext.current as ComponentActivity
    val updateViewModel: AppUpdateViewModel = hiltViewModel(viewModelStoreOwner = activity)
    AppUpdateHost(updateViewModel)

    val loggedIn = hasToken ?: return
    val startDest = remember { if (loggedIn) Routes.MAIN else Routes.LOGIN }

    val navController = rememberNavController()
    val navigator = remember(navController, mainViewModel) { AppNavigator(navController, mainViewModel) }
    val textToolbar = LocalTextToolbar.current
    DisposableEffect(navController) {
        val listener = NavController.OnDestinationChangedListener { _, _, _ -> textToolbar.hide() }
        navController.addOnDestinationChangedListener(listener)
        onDispose { navController.removeOnDestinationChangedListener(listener) }
    }

    fun consumeDeepLink(target: ShareLinks.Target) {
        val resolved = mainViewModel.claimDeepLink(target) ?: return
        navController.navigate(Routes.MAIN) { launchSingleTop = true }
        navController.navigateShareTarget(resolved)
    }

    LaunchedEffect(loggedIn) {
        if (loggedIn) return@LaunchedEffect
        val current = navController.currentDestination?.route
        if (current != null && current != Routes.LOGIN) {
            navController.navigate(Routes.LOGIN) {
                popUpTo(0) { inclusive = true }
                launchSingleTop = true
            }
        }
    }

    LaunchedEffect(deepLink, loggedIn) {
        if (!loggedIn) return@LaunchedEffect
        deepLink?.let { consumeDeepLink(it) }
    }

    NavHost(navController = navController, startDestination = startDest) {
        authGraph(
            navigator = navigator,
            onLoggedIn = {
                navController.navigate(Routes.MAIN) {
                    popUpTo(Routes.LOGIN) { inclusive = true }
                }
                deepLink?.let { consumeDeepLink(it) }
            },
        )
        mainGraph(navigator)
        chatGraph(navigator)
        postGraph(navigator)
        accountGraph(navigator, updateViewModel)
        mapGraph(navigator)
    }
}
