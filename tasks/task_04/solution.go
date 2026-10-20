package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) == 0 {
		return Stats{
			Count: 0,
			Sum:   0,
			Min:   0,
			Max:   0,
		}
	}
	if len(nums) == 1 {
		return Stats{
			Count: 0,
			Sum:   0,
			Min:   0,
			Max:   0,
		}
	} else {
		stats := Stats{
			Count: len(nums) - 1,
			Sum:   0,
			Min:   nums[1] - nums[0],
			Max:   nums[1] - nums[0],
		}
		for i := range nums {
			if i > 0 {
				stats.Sum += nums[i] - nums[i-1]
				if nums[i]-nums[i-1] < stats.Min {
					stats.Min = nums[i] - nums[i-1]

				}
				if nums[i]-nums[i-1] > stats.Max {
					stats.Max = nums[i] - nums[i-1]
				}
			}
		}
		return stats
	}
}
