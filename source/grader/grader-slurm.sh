#!/bin/bash
#
#SBATCH --mail-user=mashalimay@cs.uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj2_grade 
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/mashalimay/ParallelProgramming/project-2-mshalimay/proj2/
#SBATCH --partition=debug
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=10:00

module load golang/1.19
# NOTE: the grader is only working when ran from inside the proj2/grader directory
cd grader 
go run grader.go proj2
