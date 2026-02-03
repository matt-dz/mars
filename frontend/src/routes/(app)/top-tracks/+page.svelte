<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import * as Alert from '$lib/components/ui/alert';
	import * as Popover from '$lib/components/ui/popover';
	import { Calendar } from '$lib/components/ui/calendar';
	import TrackList from '$lib/components/app/TrackList.svelte';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import { type DateValue, parseDate } from '@internationalized/date';

	let { data }: { data: PageData } = $props();

	let selectedPeriod = $derived<string>(data.period);
	let customStartDate = $derived<DateValue | undefined>(
		data.period === 'custom' ? parseDate(data.startDate.split('T')[0]) : undefined
	);
	let customEndDate = $derived<DateValue | undefined>(
		data.period === 'custom' ? parseDate(data.endDate.split('T')[0]) : undefined
	);
	let showCustom = $derived<boolean>(data.period === 'custom');

	function handlePeriodChange(value: string) {
		selectedPeriod = value;
		showCustom = value === 'custom';

		if (value !== 'custom') {
			// Navigate with new period
			// @ts-expect-error - resolve doesn't support query params in types but works at runtime
			goto(resolve(`/top-tracks?period=${value}`));
		}
	}

	function applyCustomRange() {
		if (customStartDate && customEndDate) {
			const startStr = `${customStartDate.year}-${String(customStartDate.month).padStart(2, '0')}-${String(customStartDate.day).padStart(2, '0')}`;
			const endStr = `${customEndDate.year}-${String(customEndDate.month).padStart(2, '0')}-${String(customEndDate.day).padStart(2, '0')}`;
			// @ts-expect-error - resolve doesn't support query params in types but works at runtime
			goto(resolve(`/top-tracks?period=custom&start=${startStr}&end=${endStr}`));
		}
	}

	function formatDate(date: DateValue | undefined): string {
		if (!date) return 'Pick a date';
		return `${String(date.month).padStart(2, '0')}/${String(date.day).padStart(2, '0')}/${date.year}`;
	}

	function getPeriodLabel(period: string): string {
		switch (period) {
			case 'week':
				return 'Past 7 Days';
			case 'month-to-date':
				return 'Month to Date';
			case 'year-to-date':
				return 'Year to Date';
			case 'custom':
				return 'Custom Range';
			case 'day':
			default:
				return 'Past 24 Hours';
		}
	}
</script>

<svelte:head>
	<title>Top Tracks - Mars</title>
</svelte:head>

<div class="container mx-auto max-w-4xl px-4 py-8">
	<div class="mb-8">
		<a
			href={resolve('/home')}
			class="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-primary"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 20 20"
				fill="currentColor"
				class="h-4 w-4"
			>
				<path
					fill-rule="evenodd"
					d="M17 10a.75.75 0 01-.75.75H5.612l4.158 3.96a.75.75 0 11-1.04 1.08l-5.5-5.25a.75.75 0 010-1.08l5.5-5.25a.75.75 0 111.04 1.08L5.612 9.25H16.25A.75.75 0 0117 10z"
					clip-rule="evenodd"
				/>
			</svg>
			Back to playlists
		</a>
	</div>

	<div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
		<div class="space-y-1">
			<h1 class="text-3xl font-bold tracking-tight">Top Tracks</h1>
			<p class="text-muted-foreground">
				{data.topTracks.tracks.length} track{data.topTracks.tracks.length === 1 ? '' : 's'} in {getPeriodLabel(
					selectedPeriod
				)}
			</p>
		</div>

		<div class="flex flex-col gap-3 sm:flex-row sm:items-end">
			<div class="space-y-2">
				<Label for="period-select" class="text-sm">Time Period</Label>
				<Select.Root
					type="single"
					value={selectedPeriod}
					onValueChange={(v) => v && handlePeriodChange(v)}
				>
					<Select.Trigger id="period-select" class="w-[200px]">
						{getPeriodLabel(selectedPeriod)}
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="day">Past 24 Hours</Select.Item>
						<Select.Item value="week">Past 7 Days</Select.Item>
						<Select.Item value="month-to-date">Month to Date</Select.Item>
						<Select.Item value="year-to-date">Year to Date</Select.Item>
						<Select.Item value="custom">Custom Range</Select.Item>
					</Select.Content>
				</Select.Root>
			</div>

			{#if showCustom}
				<div class="flex gap-2">
					<div class="space-y-2">
						<Label class="text-sm">Start Date</Label>
						<Popover.Root>
							<Popover.Trigger
								class="inline-flex h-9 w-[140px] items-center justify-start gap-2 rounded-md border border-input bg-transparent px-3 py-1 text-left text-sm shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									class="h-4 w-4"
								>
									<rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
									<line x1="16" y1="2" x2="16" y2="6" />
									<line x1="8" y1="2" x2="8" y2="6" />
									<line x1="3" y1="10" x2="21" y2="10" />
								</svg>
								<span class="font-normal">{formatDate(customStartDate)}</span>
							</Popover.Trigger>
							<Popover.Content class="w-auto p-0">
								<Calendar type="single" bind:value={customStartDate} />
							</Popover.Content>
						</Popover.Root>
					</div>

					<div class="space-y-2">
						<Label class="text-sm">End Date</Label>
						<Popover.Root>
							<Popover.Trigger
								class="inline-flex h-9 w-[140px] items-center justify-start gap-2 rounded-md border border-input bg-transparent px-3 py-1 text-left text-sm shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									class="h-4 w-4"
								>
									<rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
									<line x1="16" y1="2" x2="16" y2="6" />
									<line x1="8" y1="2" x2="8" y2="6" />
									<line x1="3" y1="10" x2="21" y2="10" />
								</svg>
								<span class="font-normal">{formatDate(customEndDate)}</span>
							</Popover.Trigger>
							<Popover.Content class="w-auto p-0">
								<Calendar type="single" bind:value={customEndDate} />
							</Popover.Content>
						</Popover.Root>
					</div>

					<div class="flex items-end">
						<Button
							onclick={applyCustomRange}
							disabled={!customStartDate || !customEndDate}
							class="gap-2"
						>
							Apply
						</Button>
					</div>
				</div>
			{/if}
		</div>
	</div>

	{#if data.error}
		<Alert.Root variant="destructive" class="mb-6">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				class="h-4 w-4"
			>
				<circle cx="12" cy="12" r="10" />
				<line x1="12" y1="8" x2="12" y2="12" />
				<line x1="12" y1="16" x2="12.01" y2="16" />
			</svg>
			<Alert.Title>Invalid Time Frame</Alert.Title>
			<Alert.Description>{data.error}</Alert.Description>
		</Alert.Root>
	{/if}

	{#if data.topTracks.tracks.length > 0}
		<TrackList tracks={data.topTracks.tracks} />
	{:else}
		<div
			class="flex flex-col items-center justify-center gap-4 rounded-2xl border border-dashed border-primary/20 bg-gradient-to-br from-primary/5 via-transparent to-destructive/5 p-16 text-center"
		>
			<div class="rounded-full bg-gradient-to-br from-primary/20 to-destructive/10 p-4">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
					class="h-8 w-8 text-primary/60"
				>
					<path d="M9 18V5l12-2v13" />
					<circle cx="6" cy="18" r="3" />
					<circle cx="18" cy="16" r="3" />
				</svg>
			</div>
			<div class="space-y-1">
				<p class="font-medium">No tracks found</p>
				<p class="text-sm text-muted-foreground">
					Try selecting a different time period or sync your Spotify listening history.
				</p>
			</div>
		</div>
	{/if}
</div>
