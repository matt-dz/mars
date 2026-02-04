<script lang="ts">
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';

	let {
		period = $bindable('all'),
		sortOrder = $bindable('desc'),
		playlistType = $bindable('all')
	}: {
		period: 'week' | 'month' | 'year' | 'all';
		sortOrder: 'asc' | 'desc';
		playlistType: 'all' | 'weekly' | 'monthly' | 'custom';
	} = $props();
</script>

<div class="flex flex-wrap items-center gap-4">
	<div class="flex items-center gap-2">
		<Label for="type-filter" class="text-sm text-muted-foreground">Type:</Label>
		<Select.Root type="single" bind:value={playlistType}>
			<Select.Trigger id="type-filter" class="w-32">
				{playlistType === 'all'
					? 'All'
					: playlistType === 'weekly'
						? 'Weekly'
						: playlistType === 'monthly'
							? 'Monthly'
							: 'Custom'}
			</Select.Trigger>
			<Select.Content>
				<Select.Item value="all">All</Select.Item>
				<Select.Item value="weekly">Weekly</Select.Item>
				<Select.Item value="monthly">Monthly</Select.Item>
				<Select.Item value="custom">Custom</Select.Item>
			</Select.Content>
		</Select.Root>
	</div>

	<div class="flex items-center gap-2">
		<Label for="period-filter" class="text-sm text-muted-foreground">Created:</Label>
		<Select.Root type="single" bind:value={period}>
			<Select.Trigger id="period-filter" class="w-32">
				{period === 'all'
					? 'All Time'
					: period === 'week'
						? 'This Week'
						: period === 'month'
							? 'This Month'
							: 'This Year'}
			</Select.Trigger>
			<Select.Content>
				<Select.Item value="all">All Time</Select.Item>
				<Select.Item value="week">This Week</Select.Item>
				<Select.Item value="month">This Month</Select.Item>
				<Select.Item value="year">This Year</Select.Item>
			</Select.Content>
		</Select.Root>
	</div>

	<div class="flex items-center gap-2">
		<Label for="sort-filter" class="text-sm text-muted-foreground">Sort:</Label>
		<Select.Root type="single" bind:value={sortOrder}>
			<Select.Trigger id="sort-filter" class="w-32">
				{sortOrder === 'desc' ? 'Newest' : 'Oldest'}
			</Select.Trigger>
			<Select.Content>
				<Select.Item value="desc">Newest</Select.Item>
				<Select.Item value="asc">Oldest</Select.Item>
			</Select.Content>
		</Select.Root>
	</div>
</div>
