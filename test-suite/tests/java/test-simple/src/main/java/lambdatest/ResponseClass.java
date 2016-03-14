package lambdatest;

public class ResponseClass {
    String greetings;
    String greetings2 = "123";
    String greetings3 = "123";

    public String getGreetings() {
        return greetings;
    }
    public String getGreetings3() {
        return greetings3;
    }

    public void setGreetings(String greetings) {
        this.greetings = greetings;
    }
    public void setGreetings3(String greetings) {
        this.greetings3 = greetings;
    }

    public ResponseClass(String greetings) {
        this.greetings = greetings;
    }

    public ResponseClass() {
    }

}
