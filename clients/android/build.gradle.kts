plugins {
    kotlin("jvm") version "2.0.21"
    jacoco
}

repositories {
    mavenCentral()
}

kotlin {
    jvmToolchain(17)
    sourceSets {
        test {
            kotlin.srcDir("../../tests/android")
        }
    }
}

dependencies {
    testImplementation(kotlin("test-junit5"))
    testRuntimeOnly("org.junit.jupiter:junit-jupiter-engine:5.11.3")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher:1.11.3")
}

tasks.test {
    useJUnitPlatform()
    finalizedBy(tasks.jacocoTestReport)
}

tasks.jacocoTestReport {
    dependsOn(tasks.test)
    reports {
        xml.required.set(true)
        html.required.set(true)
    }
}

tasks.jacocoTestCoverageVerification {
    dependsOn(tasks.test)
    violationRules {
        rule {
            element = "CLASS"
            for (counterType in listOf("LINE", "BRANCH", "INSTRUCTION", "METHOD")) {
                limit {
                    counter = counterType
                    value = "COVEREDRATIO"
                    minimum = "1.0".toBigDecimal()
                }
            }
        }
    }
}

tasks.check {
    dependsOn(tasks.jacocoTestCoverageVerification)
}
