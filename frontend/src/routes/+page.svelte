<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { resolve } from '$app/paths';
</script>

<div class="flex min-h-screen flex-col">
	<!-- Navigation -->
	<header class="border-b border-border/40 bg-background/80 backdrop-blur-sm">
		<nav class="container mx-auto flex h-16 items-center justify-between px-4">
			<a href={resolve('/')} class="flex items-center gap-2">
				<div
					class="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-primary via-primary/80 to-destructive/60"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="currentColor"
						class="h-4 w-4 text-primary-foreground"
					>
						<circle cx="12" cy="12" r="10" />
					</svg>
				</div>
				<span class="text-xl font-bold">Mars</span>
			</a>
			<div class="flex items-center gap-4">
				<a
					href="https://github.com/matt-dz/mars"
					target="_blank"
					rel="noopener noreferrer"
					aria-label="GitHub"
					class="text-muted-foreground transition-colors hover:text-foreground"
				>
					<svg viewBox="0 0 24 24" class="h-5 w-5" fill="currentColor">
						<path
							d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
						/>
					</svg>
				</a>
				<Button href={resolve('/login')} variant="ghost">Log in</Button>
				<Button href={resolve('/login')}>Get Started</Button>
			</div>
		</nav>
	</header>

	<!-- Hero Section -->
	<main class="flex-1">
		<section class="relative overflow-hidden py-24 sm:py-32">
			<!-- Background gradient -->
			<div
				class="absolute inset-0 -z-10 bg-gradient-to-br from-primary/5 via-transparent to-destructive/5"
			></div>
			<!-- Decorative stars -->
			<div class="pointer-events-none absolute inset-0 -z-10 overflow-hidden opacity-40">
				<div class="absolute top-[20%] left-[10%] h-1 w-1 rounded-full bg-primary"></div>
				<div
					class="absolute top-[10%] left-[25%] h-0.5 w-0.5 rounded-full bg-muted-foreground"
				></div>
				<div class="absolute top-[15%] left-[80%] h-1.5 w-1.5 rounded-full bg-primary/60"></div>
				<div
					class="absolute top-[30%] left-[70%] h-0.5 w-0.5 rounded-full bg-muted-foreground"
				></div>
				<div class="absolute top-[60%] left-[15%] h-1 w-1 rounded-full bg-primary/40"></div>
				<div
					class="absolute top-[50%] left-[90%] h-0.5 w-0.5 rounded-full bg-muted-foreground"
				></div>
				<div class="absolute top-[70%] left-[50%] h-1 w-1 rounded-full bg-primary/50"></div>
				<div
					class="absolute top-[45%] left-[5%] h-0.5 w-0.5 rounded-full bg-muted-foreground"
				></div>
			</div>

			<div class="container mx-auto px-4 text-center">
				<!-- Mars icon -->
				<div class="mb-8 flex justify-center">
					<div
						class="flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-primary/20 via-accent/15 to-destructive/10 p-2"
					>
						<div
							class="flex h-full w-full items-center justify-center rounded-full bg-gradient-to-br from-primary via-primary/80 to-destructive/60 shadow-lg shadow-primary/20"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								viewBox="0 0 24 24"
								fill="currentColor"
								class="h-8 w-8 text-primary-foreground"
							>
								<circle cx="12" cy="12" r="10" />
							</svg>
						</div>
					</div>
				</div>

				<h1
					class="mx-auto max-w-4xl text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl lg:text-7xl"
				>
					Your listening history,
					<span
						class="bg-gradient-to-r from-primary to-destructive/80 bg-clip-text text-transparent"
					>
						beautifully organized
					</span>
				</h1>
				<p class="mx-auto mt-6 max-w-2xl text-lg text-muted-foreground sm:text-xl">
					Mars automatically creates weekly and monthly playlists from your Spotify listening
					history. Rediscover your favorite tracks and see how your taste evolves over time.
				</p>
				<div class="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
					<Button href={resolve('/login')} size="lg" class="gap-2">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 24 24"
							fill="currentColor"
							class="h-5 w-5"
						>
							<path
								d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"
							/>
						</svg>
						Connect with Spotify
					</Button>
					<Button
						href="https://github.com/matt-dz/mars"
						target="_blank"
						rel="noopener noreferrer"
						variant="outline"
						size="lg"
						class="gap-2"
					>
						<svg viewBox="0 0 24 24" class="h-5 w-5" fill="currentColor">
							<path
								d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
							/>
						</svg>
						View on GitHub
					</Button>
				</div>
			</div>
		</section>

		<!-- Features Section -->
		<section class="border-t border-border/40 py-24">
			<div class="container mx-auto px-4">
				<div class="mb-16 text-center">
					<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
						Everything happens automatically
					</h2>
					<p class="mt-4 text-lg text-muted-foreground">
						Connect your Spotify account and let Mars do the rest.
					</p>
				</div>

				<div class="grid gap-8 md:grid-cols-3">
					<!-- Feature 1 -->
					<div
						class="group rounded-2xl border border-border/50 bg-gradient-to-br from-card to-card/50 p-8 transition-all hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5"
					>
						<div
							class="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-primary/20 to-primary/5"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								class="h-6 w-6 text-primary"
							>
								<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
								<line x1="16" y1="2" x2="16" y2="6"></line>
								<line x1="8" y1="2" x2="8" y2="6"></line>
								<line x1="3" y1="10" x2="21" y2="10"></line>
							</svg>
						</div>
						<h3 class="mb-2 text-xl font-semibold">Weekly & Monthly Playlists</h3>
						<p class="text-muted-foreground">
							Automatically generated playlists based on your listening habits. See your top tracks
							for each week and month.
						</p>
					</div>

					<!-- Feature 2 -->
					<div
						class="group rounded-2xl border border-border/50 bg-gradient-to-br from-card to-card/50 p-8 transition-all hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5"
					>
						<div
							class="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-primary/20 to-primary/5"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								class="h-6 w-6 text-primary"
							>
								<path d="M12 20V10"></path>
								<path d="M18 20V4"></path>
								<path d="M6 20v-4"></path>
							</svg>
						</div>
						<h3 class="mb-2 text-xl font-semibold">Track Your Plays</h3>
						<p class="text-muted-foreground">
							See exactly how many times you've listened to each track. Discover patterns in your
							music taste.
						</p>
					</div>

					<!-- Feature 3 -->
					<div
						class="group rounded-2xl border border-border/50 bg-gradient-to-br from-card to-card/50 p-8 transition-all hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5"
					>
						<div
							class="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-primary/20 to-primary/5"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								viewBox="0 0 24 24"
								fill="currentColor"
								class="h-6 w-6 text-primary"
							>
								<path
									d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"
								/>
							</svg>
						</div>
						<h3 class="mb-2 text-xl font-semibold">Sync to Spotify</h3>
						<p class="text-muted-foreground">
							Export your generated playlists directly to your Spotify account with a single click.
						</p>
					</div>
				</div>
			</div>
		</section>

		<!-- CTA Section -->
		<section
			class="border-t border-border/40 bg-gradient-to-br from-primary/5 via-transparent to-destructive/5 py-24"
		>
			<div class="container mx-auto px-4 text-center">
				<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">Ready to explore your music?</h2>
				<p class="mt-4 text-lg text-muted-foreground">
					Connect your Spotify account and start discovering your listening patterns.
				</p>
				<div class="mt-8">
					<Button href={resolve('/login')} size="lg">Get Started</Button>
				</div>
			</div>
		</section>
	</main>

	<!-- Footer -->
	<footer class="border-t border-border/40 py-8">
		<div
			class="container mx-auto flex flex-col items-center justify-between gap-4 px-4 sm:flex-row"
		>
			<div class="flex items-center gap-2 text-sm text-muted-foreground">
				<div
					class="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-br from-primary via-primary/80 to-destructive/60"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="currentColor"
						class="h-2.5 w-2.5 text-primary-foreground"
					>
						<circle cx="12" cy="12" r="10" />
					</svg>
				</div>
				<span>Mars</span>
			</div>
			<div class="flex items-center gap-6 text-sm text-muted-foreground">
				<a
					href="https://github.com/matt-dz/mars"
					target="_blank"
					rel="noopener noreferrer"
					class="transition-colors hover:text-foreground"
				>
					GitHub
				</a>
			</div>
		</div>
	</footer>
</div>
