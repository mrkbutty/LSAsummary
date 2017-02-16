# summarize LSA output into more human friendly form

BEGIN {
	FS="|"
	section=0
}

function ltrim(s) { sub(/^[ \t\r\n]+/, "", s); return s }
function rtrim(s) { sub(/[ \t\r\n]+$/, "", s); return s }
function trim(s) { return rtrim(ltrim(s)); }

function round(x,   ival, aval, fraction)  {
  ival = int(x)    # integer part, int() truncates

  # see if fractional part
  if (ival == x)   # no fraction
     return ival   # ensure no decimals

  if (x < 0) {
     aval = -x     # absolute value
     ival = int(aval)
     fraction = aval - ival
     if (fraction >= .5)
        return int(x) - 1   # -2.5 --> -3
     else
        return int(x)       # -2.3 --> -2
  } else {
     fraction = x - ival
     if (fraction >= .5)
        return ival + 1
     else
        return ival
  }
}

function summary() {

	if (length(tierPages) > 1) {
		print "< Tier Sizing >"
		for (i in tierPages) {
			if (tierPages[i] > 0 && tierPotential[i] > 0) piophpp=round(tierPotential[i] / tierPages[i])
			else piophpp=0
			percent = (tierPages[i] / tiertotal) * 100
			
			printf "%13s = %-3.1f%%    Potential-IOPH-per-Page = %d\n", i, percent, piophpp
		}
		if (tiertotal > 0 && tierused > 0) pused = (tierused/tiertotal) * 100
		else pused=0
		print "      ------------------------------------------------------------"
		printf "          Total Pages = %d  Used = %d (%3.1f%%)", tiertotal, tierused, pused
		print "\n"
	}
	
	print "< Lane Summary >"
	printf "%6s %14s %14s %14s %14s %14s\n", "Policy", "Active-Pages", "Inactive-Pages", "Used-Pages", "Total-IOPH", "Access-Density"

	num=asorti(activetot,keys)
	for (i=1; i<=num; i++) {
		totpages=activetot[keys[i]]+zerotot[keys[i]]
		if ( activetot[keys[i]] > 0 && totIOPH[keys[i]] > 0 ) access = totIOPH[keys[i]] / activetot[keys[i]]
		else access=0

		printf "%6s %14d %14d %14d %14d %14.2f\n", keys[i], activetot[keys[i]], zerotot[keys[i]], totpages, totIOPH[keys[i]], access
	}
	printf "\n"
	print "< Potential IOPH by Lane & Tier >"
	printf "    Lane-Policy"
	for (j=1; j<=tiercount; j++) printf " %14s", tiername[j]
	printf "\n"
	for (i=1; i<=num; i++) {
		totpages=activetot[keys[i]]+zerotot[keys[i]]
		printf " %14s", keys[i]
		for (j=1; j<=tiercount; j++) {
			tier=tiername[j]
			if (tierPages[tier] > 0 && tierPotential[tier] > 0) piophpp=round(tierPotential[tier] / tierPages[tier])
			else piophpp=0
			pioph=totpages*piophpp
			percent=pioph/(tierPotential[tier]/100)
			printf " %13d%%", percent
			#printf "\n  %d   %d   \n", pioph, tierPotential[tier]
		}
		printf "\n"
	}
	printf "\n"
	
	delete activetot
	delete zerotot
	delete totIOPH
	delete tierPotential
	delete tierPages
	tiercount=0
	tiertotal=0
	tierused=0
}

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

