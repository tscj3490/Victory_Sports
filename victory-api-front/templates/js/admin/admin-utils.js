export function toArray(array) {
    return Array.prototype.slice.call(array);
}
export function mapExtractArrayField(fieldName) {
    return function (value, entry) {
        let resp = [];
        value.forEach(function (obj) {
            resp.push(obj[fieldName]);
        });
        return resp;
    };
}
export function mapExtractField(fieldName) {
    return function (value, entry) {
        return value[fieldName];
    }
}
export function mapExtractEntryField(fieldName) {
    return function (value, entry) {
        return entry[fieldName];
    }
}
