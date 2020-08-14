
echo "executing workflow ......."

echo "step 1 ......."

echo "step m ......."

>&2 echo "encountering an error ......"
>&2 echo "error 1 ......"
>&2 echo "error 2 ......"
exit -1

echo "step n ......."

echo "end workflow!"

#./tests/functests/mock_error.sh > /tmp/ok 2>/tmp/err