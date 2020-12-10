
################################################################################
# generate enum
echo "generate enums"
genenum -typesize=uint8 -typename=FlowType -packagename=flowtype -basedir=enum -vectortype=int
genenum -typesize=uint16 -typename=CommandID -packagename=commandid -basedir=enum -vectortype=int
genenum -typesize=uint16 -typename=NotificationID -packagename=notificationid -basedir=enum -vectortype=int
genenum -typesize=uint16 -typename=ResultCode -packagename=resultcode -basedir=enum -vectortype=int

goimports -w enum

