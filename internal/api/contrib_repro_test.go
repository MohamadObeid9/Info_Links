
package api
import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"
)
func TestReproLinkType(t *testing.T) {
    fake := &fakeContributionService{}
    h := testHandler(t, withContribution(fake))
    body := `{"course_name":"A","link_url":"https://example.com","link_type":"drive","note":""}`
    req := httptest.NewRequest(http.MethodPost, "/api/contributions", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    h.handlePostContribution(rr, req)
    t.Logf("status=%d body=%s calls=%d", rr.Code, rr.Body.String(), fake.createCalls)
    if rr.Code != http.StatusCreated {
        t.Fatalf("expected 201 got %d", rr.Code)
    }
}
