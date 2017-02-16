# summarize LSA output into more human friendly form

BEGIN {
	FS="|"
	section=0
}

function ltrim(s) { sub(/^[ \t\r\n]+/, "", s); return s }
function rtrim(s) { sub(/[ \t\r\n]+$/, "", s); return s }
function trim(s) { return rtrim(ltrim(s)); }


/^[ \s]*$/ { 
	section=0
	print
}   # reset on blank line

/^ *Efficiency|^DRDSTS|^CVDEV TYPE|^Usage/ { section=1 }  # Key guides

/.*/ {
	if (section==1) {
		print
		next
	}
}


/^\|CVDEV#\||POOL\#|DIR\#/ {
	section=2

	header[2]=trim($2)
	headstr=header[2]
	for(i = 3; i < NF; i++) {
		header[i]=trim($i)
		gsub(/\(BLK\)/, "(TB)", header[i])
		gsub(/\(MB\)/, "(GB)", header[i])
		headstr=headstr "," header[i]
	}
	print headstr
	next
}

/^\|----+\|/ { next }

/.*/ {
	if (section == 2) {
		datastr=trim($2)
		for(i = 3; i < NF; i++) {
			item=trim($i)
			if (index(header[i], "(TB)") > 0) {  # It was in hex BLK and needs coverting to TB
				match(item, /(0x[0-9ABCDEF]+)(\[.*\])?/, result)
				#print result[1] result[2]
				item=strtonum(result[1])
				if (item > 0) { item=(item *512)/1024/1024/1024/1024 }
				item=sprintf("%.2f%s", item, result[2])
			}
			if (index(header[i], "(GB)") > 0) {  # It was in MB and needs coverting to GB
				match(item, /(0x[0-9ABCDEF]+)(\[.*\])?/, result)
				if (item > 0) { item=item/1024 }
				item=sprintf("%.1f", item)
			}
			datastr=datastr "," item

		}
		print datastr
	}
	
}

