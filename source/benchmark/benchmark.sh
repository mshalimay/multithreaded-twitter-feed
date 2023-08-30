#!/bin/bash
#
#SBATCH --mail-user=mashalimay@cs.uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj2_benchmark 
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/mashalimay/ParallelProgramming/project-2-mshalimay/proj2/
#SBATCH --partition=debug 
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=1:00:00


module load golang/1.19
go mod tidy

testSizes=("xsmall" "small" "medium" "large" "xlarge") 
n_threads=(1 2 4 6 8 10 12)   
repeat=5       # number of times to repeat each combination of testSize x n_thread
resultsFile="./benchmark/results.txt" # file to output resulting elapsed times

# clean results file
echo "" > $resultsFile

# loop through all test sizes and threads
for testSize in ${testSizes[@]}
do
    for n_thread in ${n_threads[@]}
    do
        if [ "$n_thread" == "1" ]
        then
            version="s"
        else
            version="p"
        fi

        # run the benchmark `repeat` times
        for (( i=0; i<$repeat; i++ ))
        do
            if [ "$n_thread" == "1" ]
            then
                output=$(go run ./benchmark/benchmark.go "s" "$testSize")
            else
                output=$(go run ./benchmark/benchmark.go "p" "$testSize" "$n_thread")
            fi

            if [ $? -ne 0 ]
            then
                echo "Failed to start command"
                exit 1
            fi

            echo "{\"version\":\"$version\", \"testSize\":\"$testSize\", \"threads\":$n_thread, \"time\":$output}" >> $resultsFile
        done
    done
done

go run ./plotter/plot.go