# Golang-Challenge

## Decisions Taken
* Save the creation date next to the item price to be able to compare the `maxage` with that date.
* Mutex was added to the cache to avoid concurrent writes of the goroutines.
* Parallelization of the getPrice(...) calls through a select pattern.
* Use of two different channels to manage Parallelization. One for prices and the other for errors. In case of any error the asynchronous execution ends with an error.
* Timeout of two seconds for each parallel call to avoid deadlock or any type of leak
* In all structs the pointer to the nested structs is saved and not their values to improve performance.
