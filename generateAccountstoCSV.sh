#!/bin/bash

# https://github.com/aliirns

#          ,_---~~~~~----._
#   _,,_,*^____      _____``*g*\"*,
#  / __/ /'     ^.  /      \ ^@q   f
# [  @f | @))    |  | @))   l  0 _/
#  \`/   \~____ / __ \_____/    \
#   |           _l__l_           I
#   }          [______]           I
#   ]            | | |            |
#   ]             ~ ~             |
#   |                            |
#    |                           |


filename="TestAccounts.csv"
prefix="loadTestAccount"

echo "running..."

for((i=$1; i<$2; i++)); do
    
    pylonsd keys add $prefix$i &> /dev/null
    #writing to file
    address=$(pylonsd keys show $prefix$i | grep -o 'pylo[a-z,0-9]*')
    privKey=$(yes | pylonsd keys export $prefix$i --unsafe --unarmored-hex 2> file.txt && tail -n 1 file.txt)
    rm file.txt
    echo $prefix$i,$address,$privKey >> $filename; \
    #end writing to file
    pylonsd tx pylons create-account $prefix$i "" "" --from $prefix$i --yes &> /dev/null
    sleep 0.5
    
    
    
done


echo "Exiting..."


